import { ApiError } from "./ApiError";

import type { ZodError } from "zod";

export function handleZodError<T>(error: ZodError<T>) {
  const formattedZodError = error.issues.map((err) => ({
    field: err.path.join("."),
    message: err.message,
  }));
  return {
    message: "Validation Failed",
    statusCode: 400,
    errors: formattedZodError,
  };
}

export class ZodCustomError<T = unknown> extends ApiError {
  constructor(error: ZodError<T>) {
    super(handleZodError(error));
  }
}
