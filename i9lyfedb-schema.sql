--
-- PostgreSQL database dump
--

\restrict mOGNznRHC75IERujSBVq5pbQkfbpY6bHE6hdQc82NOEKBKcMhyraN9ksa4HYUMV

-- Dumped from database version 18.0 (Debian 18.0-1.pgdg13+3)
-- Dumped by pg_dump version 18.1 (Ubuntu 18.1-1.pgdg24.04+2)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: message_struct; Type: TYPE; Schema: public; Owner: i9
--

CREATE TYPE public.message_struct AS (
	id_ uuid,
	che_type text,
	content_ json,
	delivery_status text,
	created_at bigint,
	sender json,
	reply_target_msg json,
	ffu boolean,
	ftu boolean
);


ALTER TYPE public.message_struct OWNER TO i9;

--
-- Name: msg_reaction_struct; Type: TYPE; Schema: public; Owner: i9
--

CREATE TYPE public.msg_reaction_struct AS (
	che_id uuid,
	che_type text,
	emoji text,
	reactor json,
	to_msg_id uuid
);


ALTER TYPE public.msg_reaction_struct OWNER TO i9;

--
-- Name: ack_msg(text, text, uuid, text, bigint); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.ack_msg(from_user text, to_user text, msg_id uuid, ack_val text, at_val bigint) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
  msg_received_in_chat bool;
BEGIN

SELECT EXISTS (SELECT true FROM chat_history_in_chat 
WHERE owner_user = from_user AND partner_user = to_user AND che_id = msg_id AND receipt = 'received')
INTO msg_received_in_chat;

IF NOT msg_received_in_chat THEN
  RETURN false;
END IF;

IF ack_val = 'delivered' THEN
  UPDATE chat_history_entry
  SET delivery_status = ack_val, delivered_at = at_val
  WHERE id_ = msg_id AND type_ = 'message' AND delivery_status = 'sent';
  IF FOUND THEN
    RETURN true;
  END IF;
ELSIF ack_val = 'read' THEN
  UPDATE chat_history_entry
  SET delivery_status = ack_val, read_at = at_val
  WHERE id_ = msg_id AND type_ = 'message' AND delivery_status IN ('sent', 'delivered');
  IF FOUND THEN
    RETURN true;
  END IF;
END IF;

RETURN false;
END;$$;


ALTER FUNCTION public.ack_msg(from_user text, to_user text, msg_id uuid, ack_val text, at_val bigint) OWNER TO i9;

--
-- Name: delete_msg(text, text, uuid, text, bigint); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.delete_msg(from_user text, to_user text, msg_id uuid, deletefor text, at_val bigint) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN

  IF deletefor = 'everyone' THEN
    UPDATE chat_history_in_chat
	SET deleted = true, deleted_at = at_val
	WHERE che_id = msg_id AND (
	  owner_user = from_user AND partner_user = to_user AND receipt = 'sent' 
	  OR 
	  owner_user = to_user AND partner_user = from_user AND receipt = 'received'
	);
	IF FOUND THEN
	  RETURN true;
	END IF;
  ELSIF deletefor = 'me' THEN
    UPDATE chat_history_in_chat
	SET deleted = true, deleted_at = at_val
	WHERE che_id = msg_id AND owner_user = from_user AND partner_user = to_user AND receipt IN ('sent', 'received');
	IF FOUND THEN
	  RETURN true;
	END IF;
  END IF;

  RETURN false;
END;
$$;


ALTER FUNCTION public.delete_msg(from_user text, to_user text, msg_id uuid, deletefor text, at_val bigint) OWNER TO i9;

--
-- Name: react_to_msg(text, text, uuid, text, bigint); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.react_to_msg(from_user text, to_user text, msg_id uuid, emoji_val text, at_val bigint) RETURNS public.msg_reaction_struct
    LANGUAGE plpgsql
    AS $$
DECLARE
  msg_in_chat bool;
  che_id_val uuid;
  reactor_user json;
BEGIN

SELECT EXISTS (SELECT 1 FROM chat_history_in_chat 
WHERE owner_user = from_user AND partner_user = to_user AND che_id = msg_id AND (
  SELECT type_ FROM chat_history_entry WHERE id_ = msg_id) = 'message'
)
INTO msg_in_chat;

IF NOT msg_in_chat THEN
  RETURN null;
END IF;

INSERT INTO chat_history_entry (type_, reactor_username, emoji, reaction_at, reaction_to)
VALUES ('reaction', from_user, emoji_val, at_val, msg_id)
RETURNING id_ INTO che_id_val;

INSERT INTO chat_history_in_chat (owner_user, partner_user, che_id, receipt)
VALUES (from_user, to_user, che_id_val, 'sent');

INSERT INTO chat_history_in_chat (owner_user, partner_user, che_id, receipt)
VALUES (to_user, from_user, che_id_val, 'received');

SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url)
FROM users WHERE username = from_user
INTO reactor_user;
  
RETURN ROW(che_id_val, 'reaction', emoji_val, reactor_user, msg_id)::msg_reaction_struct;

END;
$$;


ALTER FUNCTION public.react_to_msg(from_user text, to_user text, msg_id uuid, emoji_val text, at_val bigint) OWNER TO i9;

--
-- Name: remove_msg_reaction(text, text, uuid); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.remove_msg_reaction(from_user text, to_user text, msg_id uuid) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE 
  msg_in_chat bool;
  che_id_val uuid;
BEGIN
  SELECT EXISTS (SELECT 1 FROM chat_history_in_chat 
  WHERE owner_user = from_user AND partner_user = to_user AND che_id = msg_id AND (
    SELECT type_ FROM chat_history_entry WHERE id_ = msg_id) = 'message'
  )
  INTO msg_in_chat;

  IF NOT msg_in_chat THEN
    RETURN '';
  END IF;
		
  DELETE FROM chat_history_entry
  WHERE reactor_username = from_user AND reaction_to = msg_id
  RETURNING id_ INTO che_id_val;

  RETURN che_id_val;
END;
$$;


ALTER FUNCTION public.remove_msg_reaction(from_user text, to_user text, msg_id uuid) OWNER TO i9;

--
-- Name: reply_to_msg(text, text, json, bigint, uuid); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.reply_to_msg(from_user text, to_user text, content_val json, created_at_val bigint, reply_target_msg_id uuid) RETURNS public.message_struct
    LANGUAGE plpgsql
    AS $$
DECLARE
  msg_in_chat bool;
  che_id_val uuid;
  sender_user json;
  reply_target_msg json;
BEGIN

SELECT EXISTS (SELECT 1 FROM chat_history_in_chat 
WHERE owner_user = from_user AND partner_user = to_user AND che_id = reply_target_msg_id AND (
  SELECT type_ FROM chat_history_entry WHERE id_ = reply_target_msg_id) = 'message'
)
INTO msg_in_chat;

IF NOT msg_in_chat THEN
  RETURN null;
END IF;

INSERT INTO chat_history_entry (type_, content_, sender_username, delivery_status, created_at, reply_to)
VALUES ('message', content_val, from_user, 'sent', created_at_val, reply_target_msg_id)
RETURNING id_ INTO che_id_val;

INSERT INTO chat_history_in_chat (owner_user, partner_user, che_id, receipt)
VALUES (from_user, to_user, che_id_val, 'sent');
  
INSERT INTO chat_history_in_chat (owner_user, partner_user, che_id, receipt)
VALUES (to_user, from_user, che_id_val, 'received');

SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url, 'presence', presence)
FROM users WHERE username = from_user
INTO sender_user;

SELECT json_build_object('id', id_, 'content', content_, 'sender_username', sender_username)
FROM chat_history_entry WHERE id_ = reply_target_msg_id
INTO reply_target_msg;

  
RETURN ROW(che_id_val, 'message', content_val, 'sent', created_at_val, sender_user, reply_target_msg, false, false)::message_struct;

END;
$$;


ALTER FUNCTION public.reply_to_msg(from_user text, to_user text, content_val json, created_at_val bigint, reply_target_msg_id uuid) OWNER TO i9;

--
-- Name: send_message(text, text, json, bigint); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.send_message(from_user text, to_user text, content_val json, created_at_val bigint) RETURNS public.message_struct
    LANGUAGE plpgsql
    AS $$
DECLARE
  che_id_val uuid;
  sender_user json;
  first_from_user boolean := false;
  first_to_user boolean := false;
BEGIN

  IF NOT EXISTS (SELECT 1 FROM user_chats_user WHERE owner_user = from_user AND partner_user = to_user) THEN
	first_from_user := true;
	
    INSERT INTO user_chats_user (owner_user, partner_user)
    VALUES (from_user, to_user);
  END IF;

  IF NOT EXISTS (SELECT 1 FROM user_chats_user WHERE owner_user = to_user AND partner_user = from_user) THEN
	first_to_user := true;
	
    INSERT INTO user_chats_user (owner_user, partner_user)
    VALUES (to_user, from_user);
  END IF;

  INSERT INTO chat_history_entry (type_, content_, sender_username, delivery_status, created_at)
  VALUES ('message', content_val, from_user, 'sent', created_at_val)
  RETURNING id_ INTO che_id_val;

  INSERT INTO chat_history_in_chat (owner_user, partner_user, che_id, receipt)
  VALUES (from_user, to_user, che_id_val, 'sent');
  
  INSERT INTO chat_history_in_chat (owner_user, partner_user, che_id, receipt)
  VALUES (to_user, from_user, che_id_val, 'received');

  SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url, 'presence', presence)
  FROM users WHERE username = from_user
  INTO sender_user;
  
  RETURN ROW(che_id_val, 'message', content_val, 'sent', created_at_val, sender_user, null, first_from_user, first_to_user)::message_struct;
END;
$$;


ALTER FUNCTION public.send_message(from_user text, to_user text, content_val json, created_at_val bigint) OWNER TO i9;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: chat_history_entry; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.chat_history_entry (
    id_ uuid DEFAULT gen_random_uuid() NOT NULL,
    type_ text NOT NULL,
    content_ jsonb,
    sender_username text,
    delivery_status text,
    reply_to uuid,
    reactor_username text,
    emoji text,
    reaction_to uuid,
    delivered_at bigint,
    read_at bigint,
    edited_at bigint,
    created_at bigint,
    reaction_at bigint,
    CONSTRAINT chat_history_entry_delivery_status_check CHECK ((delivery_status = ANY (ARRAY['sent'::text, 'delivered'::text, 'read'::text]))),
    CONSTRAINT chat_history_entry_type__check CHECK ((type_ = ANY (ARRAY['message'::text, 'reaction'::text])))
);


ALTER TABLE public.chat_history_entry OWNER TO i9;

--
-- Name: chat_history_in_chat; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.chat_history_in_chat (
    owner_user text NOT NULL,
    partner_user text NOT NULL,
    che_id uuid NOT NULL,
    receipt text NOT NULL,
    deleted boolean,
    deleted_at bigint,
    CONSTRAINT chat_history_in_chat_receipt_check CHECK ((receipt = ANY (ARRAY['sent'::text, 'received'::text])))
);


ALTER TABLE public.chat_history_in_chat OWNER TO i9;

--
-- Name: comment_mentions_user; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.comment_mentions_user (
    comment_id uuid NOT NULL,
    username text NOT NULL
);


ALTER TABLE public.comment_mentions_user OWNER TO i9;

--
-- Name: hashtags; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.hashtags (
    htname text NOT NULL
);


ALTER TABLE public.hashtags OWNER TO i9;

--
-- Name: post_includes_hashtag; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.post_includes_hashtag (
    post_id uuid NOT NULL,
    htname text NOT NULL
);


ALTER TABLE public.post_includes_hashtag OWNER TO i9;

--
-- Name: post_mentions_user; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.post_mentions_user (
    post_id uuid NOT NULL,
    username text NOT NULL
);


ALTER TABLE public.post_mentions_user OWNER TO i9;

--
-- Name: posts; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.posts (
    id_ uuid DEFAULT gen_random_uuid() NOT NULL,
    owner_user text NOT NULL,
    type_ text NOT NULL,
    media_cloud_names text[] NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    deleted boolean DEFAULT false,
    reposted_by_user text,
    created_at bigint,
    deleted_at bigint,
    CONSTRAINT posts_type__check CHECK ((type_ = ANY (ARRAY['photo'::text, 'video'::text, 'reel'::text])))
);


ALTER TABLE public.posts OWNER TO i9;

--
-- Name: user_chats_user; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.user_chats_user (
    owner_user text NOT NULL,
    partner_user text NOT NULL
);


ALTER TABLE public.user_chats_user OWNER TO i9;

--
-- Name: user_comments_on; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.user_comments_on (
    comment_id uuid DEFAULT gen_random_uuid() NOT NULL,
    username text NOT NULL,
    parent_comment_id uuid,
    post_id uuid,
    comment_text text NOT NULL,
    attachment_cloud_name text NOT NULL,
    deleted boolean DEFAULT false,
    deleted_at bigint,
    at_ bigint,
    CONSTRAINT on_post_xor_on_comment CHECK ((((post_id IS NULL) AND (parent_comment_id IS NOT NULL)) OR ((post_id IS NOT NULL) AND (parent_comment_id IS NULL))))
);


ALTER TABLE public.user_comments_on OWNER TO i9;

--
-- Name: user_follows_user; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.user_follows_user (
    follower_username text NOT NULL,
    following_username text NOT NULL,
    at_ bigint
);


ALTER TABLE public.user_follows_user OWNER TO i9;

--
-- Name: user_reacts_to_comment; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.user_reacts_to_comment (
    username text NOT NULL,
    comment_id uuid NOT NULL,
    emoji text NOT NULL,
    at_ bigint
);


ALTER TABLE public.user_reacts_to_comment OWNER TO i9;

--
-- Name: user_reacts_to_post; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.user_reacts_to_post (
    username text NOT NULL,
    post_id uuid NOT NULL,
    emoji text NOT NULL,
    at_ bigint
);


ALTER TABLE public.user_reacts_to_post OWNER TO i9;

--
-- Name: user_saves_post; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.user_saves_post (
    username text NOT NULL,
    post_id uuid NOT NULL
);


ALTER TABLE public.user_saves_post OWNER TO i9;

--
-- Name: users; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.users (
    username text NOT NULL,
    email text NOT NULL,
    password_ text NOT NULL,
    name_ text NOT NULL,
    bio text DEFAULT 'Thanks for using i9lyfe'::text,
    profile_pic_url text DEFAULT '{notset}'::text,
    presence text DEFAULT 'online'::text NOT NULL,
    birthday bigint,
    last_seen bigint,
    CONSTRAINT users_presence_check CHECK ((presence = ANY (ARRAY['online'::text, 'offline'::text])))
);


ALTER TABLE public.users OWNER TO i9;

--
-- Name: chat_history_entry chat_history_entry_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat_history_entry
    ADD CONSTRAINT chat_history_entry_pkey PRIMARY KEY (id_);


--
-- Name: hashtags hashtags_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.hashtags
    ADD CONSTRAINT hashtags_pkey PRIMARY KEY (htname);


--
-- Name: comment_mentions_user no_dup_comment_ment; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.comment_mentions_user
    ADD CONSTRAINT no_dup_comment_ment UNIQUE (comment_id, username);


--
-- Name: user_reacts_to_comment no_dup_comment_rxn; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_reacts_to_comment
    ADD CONSTRAINT no_dup_comment_rxn UNIQUE (username, comment_id);


--
-- Name: post_includes_hashtag no_dup_htname; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.post_includes_hashtag
    ADD CONSTRAINT no_dup_htname UNIQUE (post_id, htname);


--
-- Name: chat_history_entry no_dup_msg_rxn; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat_history_entry
    ADD CONSTRAINT no_dup_msg_rxn UNIQUE (reactor_username, reaction_to);


--
-- Name: post_mentions_user no_dup_post_ment; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.post_mentions_user
    ADD CONSTRAINT no_dup_post_ment UNIQUE (post_id, username);


--
-- Name: user_reacts_to_post no_dup_post_rxn; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_reacts_to_post
    ADD CONSTRAINT no_dup_post_rxn UNIQUE (username, post_id);


--
-- Name: user_saves_post no_dup_saves; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_saves_post
    ADD CONSTRAINT no_dup_saves UNIQUE (username, post_id);


--
-- Name: posts posts_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_pkey PRIMARY KEY (id_);


--
-- Name: user_chats_user ucu_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_chats_user
    ADD CONSTRAINT ucu_pkey PRIMARY KEY (owner_user, partner_user);


--
-- Name: user_comments_on user_comments_on_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_comments_on
    ADD CONSTRAINT user_comments_on_pkey PRIMARY KEY (comment_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (username);


--
-- Name: chat_history_entry chat_history_entry_reaction_to_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat_history_entry
    ADD CONSTRAINT chat_history_entry_reaction_to_fkey FOREIGN KEY (reaction_to) REFERENCES public.chat_history_entry(id_) ON DELETE CASCADE;


--
-- Name: chat_history_entry chat_history_entry_reactor_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat_history_entry
    ADD CONSTRAINT chat_history_entry_reactor_username_fkey FOREIGN KEY (reactor_username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: chat_history_entry chat_history_entry_reply_to_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat_history_entry
    ADD CONSTRAINT chat_history_entry_reply_to_fkey FOREIGN KEY (reply_to) REFERENCES public.chat_history_entry(id_) ON DELETE CASCADE;


--
-- Name: chat_history_entry chat_history_entry_sender_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat_history_entry
    ADD CONSTRAINT chat_history_entry_sender_username_fkey FOREIGN KEY (sender_username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: chat_history_in_chat chat_history_in_chat_che_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat_history_in_chat
    ADD CONSTRAINT chat_history_in_chat_che_id_fkey FOREIGN KEY (che_id) REFERENCES public.chat_history_entry(id_) ON DELETE CASCADE;


--
-- Name: chat_history_in_chat chat_history_in_chat_owner_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat_history_in_chat
    ADD CONSTRAINT chat_history_in_chat_owner_user_fkey FOREIGN KEY (owner_user) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: chat_history_in_chat chat_history_in_chat_partner_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat_history_in_chat
    ADD CONSTRAINT chat_history_in_chat_partner_user_fkey FOREIGN KEY (partner_user) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: comment_mentions_user comment_mentions_user_comment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.comment_mentions_user
    ADD CONSTRAINT comment_mentions_user_comment_id_fkey FOREIGN KEY (comment_id) REFERENCES public.user_comments_on(comment_id) ON DELETE CASCADE;


--
-- Name: comment_mentions_user comment_mentions_user_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.comment_mentions_user
    ADD CONSTRAINT comment_mentions_user_username_fkey FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: chat_history_in_chat hist_in_chat; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat_history_in_chat
    ADD CONSTRAINT hist_in_chat FOREIGN KEY (owner_user, partner_user) REFERENCES public.user_chats_user(owner_user, partner_user) ON DELETE CASCADE;


--
-- Name: post_includes_hashtag post_includes_hashtag_htname_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.post_includes_hashtag
    ADD CONSTRAINT post_includes_hashtag_htname_fkey FOREIGN KEY (htname) REFERENCES public.hashtags(htname) ON DELETE CASCADE;


--
-- Name: post_includes_hashtag post_includes_hashtag_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.post_includes_hashtag
    ADD CONSTRAINT post_includes_hashtag_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id_) ON DELETE CASCADE;


--
-- Name: post_mentions_user post_mentions_user_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.post_mentions_user
    ADD CONSTRAINT post_mentions_user_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id_) ON DELETE CASCADE;


--
-- Name: post_mentions_user post_mentions_user_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.post_mentions_user
    ADD CONSTRAINT post_mentions_user_username_fkey FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: posts posts_owner_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_owner_user_fkey FOREIGN KEY (owner_user) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: posts posts_reposted_by_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_reposted_by_user_fkey FOREIGN KEY (reposted_by_user) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_chats_user user_chats_user_owner_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_chats_user
    ADD CONSTRAINT user_chats_user_owner_user_fkey FOREIGN KEY (owner_user) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_chats_user user_chats_user_partner_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_chats_user
    ADD CONSTRAINT user_chats_user_partner_user_fkey FOREIGN KEY (partner_user) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_comments_on user_comments_on_parent_comment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_comments_on
    ADD CONSTRAINT user_comments_on_parent_comment_id_fkey FOREIGN KEY (parent_comment_id) REFERENCES public.user_comments_on(comment_id) ON DELETE CASCADE;


--
-- Name: user_comments_on user_comments_on_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_comments_on
    ADD CONSTRAINT user_comments_on_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id_) ON DELETE CASCADE;


--
-- Name: user_comments_on user_comments_on_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_comments_on
    ADD CONSTRAINT user_comments_on_username_fkey FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_follows_user user_follows_user_follower_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_follows_user
    ADD CONSTRAINT user_follows_user_follower_username_fkey FOREIGN KEY (follower_username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_follows_user user_follows_user_following_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_follows_user
    ADD CONSTRAINT user_follows_user_following_username_fkey FOREIGN KEY (following_username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_reacts_to_comment user_reacts_to_comment_comment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_reacts_to_comment
    ADD CONSTRAINT user_reacts_to_comment_comment_id_fkey FOREIGN KEY (comment_id) REFERENCES public.user_comments_on(comment_id) ON DELETE CASCADE;


--
-- Name: user_reacts_to_comment user_reacts_to_comment_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_reacts_to_comment
    ADD CONSTRAINT user_reacts_to_comment_username_fkey FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_reacts_to_post user_reacts_to_post_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_reacts_to_post
    ADD CONSTRAINT user_reacts_to_post_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id_) ON DELETE CASCADE;


--
-- Name: user_reacts_to_post user_reacts_to_post_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_reacts_to_post
    ADD CONSTRAINT user_reacts_to_post_username_fkey FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: user_saves_post user_saves_post_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_saves_post
    ADD CONSTRAINT user_saves_post_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id_) ON DELETE CASCADE;


--
-- Name: user_saves_post user_saves_post_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_saves_post
    ADD CONSTRAINT user_saves_post_username_fkey FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict mOGNznRHC75IERujSBVq5pbQkfbpY6bHE6hdQc82NOEKBKcMhyraN9ksa4HYUMV

