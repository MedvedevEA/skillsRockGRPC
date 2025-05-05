CREATE TABLE IF NOT EXISTS public."user"
(
    user_id uuid NOT NULL DEFAULT gen_random_uuid(),
    login character varying COLLATE pg_catalog."default" NOT NULL,
    password character varying COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT user_pk PRIMARY KEY (user_id),
    CONSTRAINT user_login_unique UNIQUE (login)
);
CREATE TABLE IF NOT EXISTS public.refresh_token
(
    refresh_token_id uuid NOT NULL DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    device_code character varying COLLATE pg_catalog."default" NOT NULL,
    expiration_at timestamp with time zone NOT NULL DEFAULT now(),
    is_revoke boolean NOT NULL DEFAULT false,
    CONSTRAINT refresh_token_pk PRIMARY KEY (refresh_token_id),
    CONSTRAINT refresh_token_user_id_fk FOREIGN KEY (user_id)
        REFERENCES public."user" (user_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID
);