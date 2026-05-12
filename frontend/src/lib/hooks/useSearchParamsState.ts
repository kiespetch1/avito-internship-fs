import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router-dom";

export interface ParamCodec<T> {
  parse(raw: string | null): T;
  serialize(value: T): string | null;
}

export const stringParam = (fallback = ""): ParamCodec<string> => ({
  parse(raw) {
    return raw ?? fallback;
  },
  serialize(value) {
    return value === "" || value === fallback ? null : value;
  },
});

export const intParam = (fallback: number, min = 1): ParamCodec<number> => ({
  parse(raw) {
    if (raw === null) return fallback;
    const n = Number.parseInt(raw, 10);
    if (Number.isNaN(n) || n < min) return fallback;
    return n;
  },
  serialize(value) {
    return value === fallback ? null : String(value);
  },
});

export const boolParam = (fallback = false): ParamCodec<boolean> => ({
  parse(raw) {
    return raw === null ? fallback : raw === "true";
  },
  serialize(value) {
    return value === fallback ? null : String(value);
  },
});

export function enumParam<T extends string>(
  values: readonly T[],
  fallback: T | null = null,
): ParamCodec<T | null> {
  return {
    parse(raw) {
      if (raw === null) return fallback;
      return (values as readonly string[]).includes(raw) ? (raw as T) : fallback;
    },
    serialize(value) {
      return value === null || value === fallback ? null : value;
    },
  };
}

type AnyParamCodec = {
  parse: (raw: string | null) => unknown;
  serialize: (value: never) => string | null;
};

type SchemaValues<S> = { [K in keyof S]: S[K] extends ParamCodec<infer V> ? V : never };

export function useSearchParamsState<S extends Record<string, AnyParamCodec>>(
  schema: S,
): [SchemaValues<S>, (patch: Partial<SchemaValues<S>>) => void] {
  const [search, setSearch] = useSearchParams();

  const values = useMemo(() => {
    const out: Record<string, unknown> = {};
    for (const key of Object.keys(schema)) {
      const codec = schema[key] as unknown as ParamCodec<unknown>;
      out[key] = codec.parse(search.get(key));
    }
    return out as SchemaValues<S>;
  }, [schema, search]);

  const update = useCallback(
    (patch: Partial<SchemaValues<S>>) => {
      setSearch(
        (prev) => {
          const next = new URLSearchParams(prev);
          for (const key of Object.keys(patch)) {
            const codec = schema[key] as unknown as ParamCodec<unknown> | undefined;
            if (!codec) continue;
            const serialized = codec.serialize(patch[key as keyof SchemaValues<S>]);
            if (serialized === null) next.delete(key);
            else next.set(key, serialized);
          }
          return next;
        },
        { replace: true },
      );
    },
    [schema, setSearch],
  );

  return [values, update];
}
