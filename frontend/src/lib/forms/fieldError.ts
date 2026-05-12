type ZodLikeIssue = { message: string };

export function fieldErrorMessage(errors: readonly unknown[]): string | undefined {
  for (const e of errors) {
    if (e === undefined || e === null) continue;
    if (typeof e === "string") return e;
    if (typeof e === "object" && "message" in e) {
      const msg = (e as ZodLikeIssue).message;
      if (typeof msg === "string") return msg;
    }
  }
  return undefined;
}
