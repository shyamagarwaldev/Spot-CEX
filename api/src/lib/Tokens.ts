import jwt from "jsonwebtoken";
import type { TokenType, UserJwtPayload } from "../types/auth";
export const CreateToken = (id: string, duration: number, type: TokenType) => {
  const token = jwt.sign({ id, type }, process.env.JWT_SECRET!, {
    expiresIn: duration,
  });

  return token;
};

export const VerifyToken = (token: string) => {
  return jwt.verify(token, process.env.JWT_SECRET!) as UserJwtPayload;
};
