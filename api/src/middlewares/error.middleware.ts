import jwt from "jsonwebtoken";
import type { NextFunction, Request, Response } from "express";

import {
  ApiError,
  ConflictError,
  NotFoundError,
  ServerError,
  UnauthorisedRequestError,
} from "../lib/ApiError";
import { PrismaClientKnownRequestError } from "@prisma/client/runtime/client";

function ErrorMiddleware(
  err: unknown,
  req: Request,
  res: Response,
  _next: NextFunction,
) {
  let error;

  if (err instanceof ApiError) {
    error = err;
  } else if (err instanceof PrismaClientKnownRequestError) {
    switch (err.code) {
      case "P2002":
        error = new ConflictError("Resource already exists");
        break;
      case "P2025":
        error = new NotFoundError("Resource");
        break;
      default:
        error = new ServerError("Database error");
    }
  } else if (err instanceof jwt.TokenExpiredError) {
    error = new UnauthorisedRequestError("Token expired");
  } else if (err instanceof jwt.JsonWebTokenError) {
    error = new UnauthorisedRequestError("Invalid token");
  } else {
    const unknown = err as Error;

    error = new ServerError(unknown?.message || "Internal Server Error");

    if (unknown?.stack) {
      error.stack = unknown.stack;
    }
  }

  error.path = req.originalUrl;

  console.error({
    timestamp: new Date().toISOString(),
    method: req.method,
    path: req.originalUrl,
    status: error.statusCode,
    error: error.name,
    message: error.message,
    errors: error.errors,
    stack: error.stack,
  });

  res.status(error.statusCode).json({
    success: false,
    statusCode: error.statusCode,
    message: error.message,
    errors: error.errors,
    path: error.path,
    ...(process.env.NODE_ENV === "development" && {
      stack: error.stack,
    }),
  });
}

export default ErrorMiddleware;
