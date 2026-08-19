import { VerifyToken } from "../lib/Tokens";
import { AsyncHandler } from "../lib/AsyncHandler";
import { ForbiddenError, UnauthorisedRequestError } from "../lib/ApiError";
import { TokenType } from "../types/auth";

export const auth = AsyncHandler(async (req, res, next) => {
  const token: string =
    req.cookies.accessToken ||
    req.headers.authorization?.replace("Bearer ", "");
  if (!token) throw new UnauthorisedRequestError("Access token missing");
  let verifiedToken = VerifyToken(token);
  if (verifiedToken.type !== TokenType.ACCESS)
    throw new ForbiddenError("Invalid Access Token");
  req.userInfo = verifiedToken;
  next();
});
