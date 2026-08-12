import type { ApiResponseOptions } from "../types";

export class ApiResponse<T = unknown> {
  statusCode: number;
  data?: T;
  message: string;
  success: boolean;
  path?: string;
  metadata?: Record<string, unknown>;
  constructor({
    statusCode,
    message = "Success",
    data,
    path,
    ...props
  }: ApiResponseOptions<T>) {
    this.statusCode = statusCode;
    this.data = (data as T) || undefined;
    this.message = message;
    this.success = statusCode < 400;
    this.path = path || undefined;
    this.metadata = props;
  }
}
