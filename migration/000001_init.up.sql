CREATE TABLE IF NOT EXISTS public."user"
(
    user_id uuid NOT NULL DEFAULT gen_random_uuid(),
    login character varying COLLATE pg_catalog."default" NOT NULL,
    password character varying COLLATE pg_catalog."default" NOT NULL,
    email character varying COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT "User_pkey" PRIMARY KEY (user_id),
    CONSTRAINT user_login_unique UNIQUE (login)
);
