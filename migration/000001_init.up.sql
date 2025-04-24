CREATE TABLE IF NOT EXISTS public."user"
(
    user_id uuid NOT NULL DEFAULT gen_random_uuid(),
    login character varying COLLATE pg_catalog."default" NOT NULL,
    password character varying COLLATE pg_catalog."default" NOT NULL,
    email character varying COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT "User_pkey" PRIMARY KEY (user_id),
    CONSTRAINT user_login_unique UNIQUE (login)
);
CREATE TABLE IF NOT EXISTS public.role
(
    role_id uuid NOT NULL DEFAULT gen_random_uuid(),
    name character varying COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT role_pkey PRIMARY KEY (role_id)
);
CREATE TABLE IF NOT EXISTS public.user_role
(
    user_role_id uuid NOT NULL DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    role_id uuid NOT NULL,
    CONSTRAINT user_service_pkey PRIMARY KEY (user_role_id)
);
CREATE TABLE IF NOT EXISTS public.token_type
(
    token_type_code "char" NOT NULL,
    name character varying COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT token_type_pkey PRIMARY KEY (token_type_code)
);
CREATE TABLE IF NOT EXISTS public.token
(
    token_id uuid NOT NULL DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    device_code character varying COLLATE pg_catalog."default" NOT NULL,
    token character varying COLLATE pg_catalog."default" NOT NULL,
    token_type_code "char" NOT NULL,
    expiration_at timestamp with time zone NOT NULL,
    is_revoked boolean NOT NULL DEFAULT false,
    CONSTRAINT token_pk PRIMARY KEY (token_id),
    CONSTRAINT token_user_user_id FOREIGN KEY (user_id)
        REFERENCES public."user" (user_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID
);
