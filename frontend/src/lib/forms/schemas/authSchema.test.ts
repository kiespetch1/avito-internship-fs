import { describe, expect, it } from "vitest";
import { authCredentialsSchema } from "./authSchema";

describe("authCredentialsSchema", () => {
  it("accepts valid email and password", () => {
    const parsed = authCredentialsSchema.parse({
      email: " User@Example.COM ",
      password: "passw0rd",
    });

    expect(parsed).toEqual({ email: "User@Example.COM", password: "passw0rd" });
  });

  it("rejects invalid email", () => {
    const parsed = authCredentialsSchema.safeParse({
      email: "not-email",
      password: "passw0rd",
    });

    expect(parsed.success).toBe(false);
  });

  it("rejects weak passwords", () => {
    for (const password of ["short1", "password", "12345678"]) {
      const parsed = authCredentialsSchema.safeParse({
        email: "user@example.com",
        password,
      });

      expect(parsed.success).toBe(false);
    }
  });
});
