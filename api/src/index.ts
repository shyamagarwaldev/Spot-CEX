import app from "./app";
import type { UserJwtPayload } from "./types/auth";

declare global {
  namespace Express {
    interface Request {
      userInfo: UserJwtPayload;
    }
  }
}

const port = process.env.PORT ?? 8080;
app.listen(port, () => {
  console.log(`Application is listening at ${port}`);
});
