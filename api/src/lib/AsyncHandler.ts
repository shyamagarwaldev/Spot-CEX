import type { NextFunction, Request, Response } from "express";
import type { AsyncHandlerPropsType } from "../types";
export const AsyncHandler = (requestHandler: AsyncHandlerPropsType) => {
  return (req: Request, res: Response, next: NextFunction) => {
    Promise.resolve(requestHandler(req, res, next)).catch((err) => next(err));
  };
};
