import prisma from "../../db";
import bcrypt from "bcrypt";
import { AsyncHandler } from "../../lib/AsyncHandler";
import { BadRequestError, UnauthorisedRequestError } from "../../lib/ApiError";
import { UserSignInSchema, UserSignUpSchema } from "../../schemas/users.schema";
import { ZodCustomError } from "../../lib/ZodError";
import { CreateToken, VerifyToken } from "../../lib/Tokens";
import { ApiResponse } from "../../lib/ApiResponse";
import { TokenType } from "../../types/auth";

export const signup = AsyncHandler(async (req, res) => {
  const { data, error, success } = UserSignUpSchema.safeParse(req.body);
  if (!success) {
    throw new ZodCustomError(error);
  }

  const hashedPassword = await bcrypt.hash(data.password, 10);
  await prisma.user.create({
    data: {
      email: data.email,
      username: data.username,
      password: hashedPassword,
    },
  });

  res.status(201).json(
    new ApiResponse({
      message: "user created succesfully",
      statusCode: 201,
    }),
  );
});

export const signin = AsyncHandler(async (req, res) => {
  const { data, error, success } = UserSignInSchema.safeParse(req.body);
  if (!success) {
    throw new ZodCustomError(error);
  }
  const user = await prisma.user.findUnique({
    where: {
      email: data.email,
    },
  });
  if (!user) {
    throw new BadRequestError("Invalid email or password");
  }

  if (!(await bcrypt.compare(data.password, user.password))) {
    throw new BadRequestError("Invalid email or password");
  }

  const accessToken = CreateToken(
    user.id,
    60 * 60 * 24 * 1000,
    TokenType.ACCESS,
  );
  const refreshToken = CreateToken(
    user.id,
    7 * 60 * 60 * 24 * 1000,
    TokenType.REFRESH,
  );

  const updatedUser = await prisma.user.update({
    where: {
      id: user.id,
    },
    select: {
      id: true,
      email: true,
      username: true,
    },
    data: {
      refresh_token: refreshToken,
    },
  });

  res
    .cookie("accessToken", accessToken, {
      maxAge: 60 * 60 * 24 * 1000,
      httpOnly: true,
      // secure: true,
    })
    .cookie("refreshToken", refreshToken, {
      maxAge: 7 * 24 * 60 * 60 * 1000,
      httpOnly: true,
      //   secure: true,
    })
    .status(200)
    .json(
      new ApiResponse({
        data: updatedUser,
        statusCode: 200,
        message: "succesfully sign in",
      }),
    );
});

export const refresh = AsyncHandler(async (req, res) => {
  const refreshToken = req.cookies.refreshToken;

  if (!refreshToken) throw new BadRequestError("refresh token is required");

  const verifiedToken = VerifyToken(refreshToken);

  if (verifiedToken.type !== TokenType.REFRESH) {
    throw new UnauthorisedRequestError("invalid refresh token");
  }
  const user = await prisma.user.findUniqueOrThrow({
    where: {
      id: verifiedToken.id,
    },
  });

  if (!user.refresh_token) {
    throw new UnauthorisedRequestError("User is not logged in");
  }

  if (user.refresh_token !== refreshToken) {
    throw new UnauthorisedRequestError("invalid refresh token");
  }
  const accessToken = CreateToken(
    verifiedToken.id,
    60 * 60 * 24,
    TokenType.ACCESS,
  );
  res
    .cookie("accessToken", accessToken, {
      maxAge: 60 * 60 * 24 * 1000,
      httpOnly: true,
      // secure: true,
    })
    .status(200)
    .json(
      new ApiResponse({
        statusCode: 200,
        message: "succesfully refreshed the token",
      }),
    );
});

export const logout = AsyncHandler(async (req, res) => {
  const id = req.userInfo.id;
  if (!req.cookies.refreshToken) {
    throw new BadRequestError("refresh token is required");
  }

  await prisma.user.update({
    where: {
      id,
      refresh_token: {
        not: null,
      },
    },
    data: {
      refresh_token: null,
    },
  });
  res
    .clearCookie("refreshToken")
    .clearCookie("accessToken")
    .status(200)
    .json(
      new ApiResponse({
        message: "user successfully logout",
        statusCode: 200,
      }),
    );
});

export const me = AsyncHandler(async (req, res) => {
  const id = req.userInfo.id;
  const user = await prisma.user.findUnique({
    where: {
      id,
    },
    select: {
      email: true,
      username: true,
      id: true,
    },
  });
  res.status(200).json(
    new ApiResponse({
      message: "successfully fetched user",
      data: user,
      statusCode: 200,
    }),
  );
});
