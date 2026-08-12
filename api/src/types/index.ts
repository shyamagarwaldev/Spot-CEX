import type { NextFunction, Request, Response } from "express";

export interface ApiErrorOptions {
  message?: string;
  statusCode: number;
  stack?: string;
  errors?: any[];
  path?: string;
}
export interface ApiResponseOptions<T = unknown> extends Record<
  string,
  unknown
> {
  statusCode: number;
  data?: T;
  message?: string;
  path?: string;
}

export type ApiResponseReturnType<T = unknown> = {
  statusCode: number;
  success: boolean;
  data?: T;
  message: string;
  path?: string;
  metadata?: Record<string, unknown>;
};

export type AsyncHandlerPropsType = (
  req: Request,
  res: Response,
  next: NextFunction,
) => Promise<void>;

export type ErrorDetail = {
  field?: string;
  message: string;
};
