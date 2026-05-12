import http from "k6/http";
import { check, fail, group, sleep } from "k6";
import { Trend } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const ROLE = __ENV.ROLE || "user";

const listLatency = new Trend("assistants_list_latency", true);

export const options = {
  scenarios: {
    ramp: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "15s", target: 20 },
        { duration: "30s", target: 50 },
        { duration: "15s", target: 0 },
      ],
      gracefulRampDown: "5s",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    "http_req_duration{endpoint:assistants_list}": ["p(95)<300", "p(99)<800"],
    checks: ["rate>0.99"],
  },
};

export function setup() {
  const loginRes = http.post(
    `${BASE_URL}/dummyLogin`,
    JSON.stringify({ role: ROLE }),
    { headers: { "Content-Type": "application/json" } },
  );
  if (loginRes.status !== 200) {
    fail(`dummyLogin failed: ${loginRes.status} ${loginRes.body}`);
  }
  const token = loginRes.json("token");

  const catsRes = http.get(`${BASE_URL}/categories`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (catsRes.status !== 200) {
    fail(`categories failed: ${catsRes.status} ${catsRes.body}`);
  }
  const categories = catsRes.json("categories") || [];
  const categoryIds = categories.map((c) => c.id);

  return { token, categoryIds };
}

const SEARCH_QUERIES = ["ассистент", "код", "поиск", "анализ", "продукт"];

function pick(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

export default function (data) {
  const headers = {
    Authorization: `Bearer ${data.token}`,
    "Content-Type": "application/json",
  };
  const tag = { endpoint: "assistants_list" };

  group("plain page=1 pageSize=10", () => {
    const res = http.get(`${BASE_URL}/assistants?page=1&pageSize=10`, {
      headers,
      tags: tag,
    });
    listLatency.add(res.timings.duration);
    check(res, {
      "status 200": (r) => r.status === 200,
      "has assistants array": (r) => Array.isArray(r.json("assistants")),
      "has pagination": (r) => r.json("pagination") !== null,
    });
  });

  group("paginated pageSize=50", () => {
    const page = 1 + Math.floor(Math.random() * 3);
    const res = http.get(`${BASE_URL}/assistants?page=${page}&pageSize=50`, {
      headers,
      tags: tag,
    });
    listLatency.add(res.timings.duration);
    check(res, { "status 200": (r) => r.status === 200 });
  });

  if (data.categoryIds.length > 0) {
    group("filter by category", () => {
      const categoryId = pick(data.categoryIds);
      const res = http.get(
        `${BASE_URL}/assistants?categoryId=${categoryId}&page=1&pageSize=20`,
        { headers, tags: tag },
      );
      listLatency.add(res.timings.duration);
      check(res, { "status 200": (r) => r.status === 200 });
    });
  }

  group("text search", () => {
    const q = encodeURIComponent(pick(SEARCH_QUERIES));
    const res = http.get(`${BASE_URL}/assistants?q=${q}&page=1&pageSize=20`, {
      headers,
      tags: tag,
    });
    listLatency.add(res.timings.duration);
    check(res, { "status 200": (r) => r.status === 200 });
  });

  sleep(Math.random() * 0.5 + 0.2);
}
