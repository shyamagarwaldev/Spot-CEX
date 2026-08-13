import { email, z } from "zod";

export const UserSignUpSchema = z.object({
  username: z
    .string("username is required")
    .min(8, "username must be atleast 8 character long"),
  password: z
    .string("password is required")
    .min(8, "minimum 8 character")
    .max(20, "max 20 character"),
  email: z.email("email is required"),
});

export type UserSignUpSchemaType = z.infer<typeof UserSignUpSchema>;

export const UserSignInSchema = z.object({
  email: z.email("email is requied"),
  password: z.string("password is required"),
});

export type UserSignInSchemaType = z.infer<typeof UserSignInSchema>;
