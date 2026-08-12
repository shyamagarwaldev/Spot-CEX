import type { JwtPayload } from "jsonwebtoken";

export interface UserJwtPayload extends JwtPayload {
  id: string;
  type: TokenType;
}

export enum TokenType {
  REFRESH = "refresh",
  ACCESS = "access",
}
