--
-- PostgreSQL database dump
--

\restrict TLrEy1DfvQIg9iGPiLyMfsdNlBkvqeW0MIMKETZEDj9eUonkPIYb4YGPG4ee8aT

-- Dumped from database version 18.4 (Debian 18.4-1.pgdg13+1)
-- Dumped by pg_dump version 18.4 (Ubuntu 18.4-1.pgdg24.04+1)

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
-- Name: che_struct; Type: TYPE; Schema: public; Owner: i9ine
--

CREATE TYPE public.che_struct AS (
	type_ text,
	id_ uuid,
	content_ jsonb,
	delivery_status text,
	created_at bigint,
	delivered_at bigint,
	read_at bigint,
	sender json,
	reply_target_msg json,
	reactor json,
	emoji text,
	rxn_to_msg json,
	cursor_ bigint
);


ALTER TYPE public.che_struct OWNER TO i9ine;

--
-- Name: comment_struct; Type: TYPE; Schema: public; Owner: i9ine
--

CREATE TYPE public.comment_struct AS (
	comment_id uuid,
	owner_user json,
	attachment_url text,
	comment_text text,
	at_ bigint,
	reactions_count integer,
	comments_count integer,
	me_reaction text,
	cursor_ bigint
);


ALTER TYPE public.comment_struct OWNER TO i9ine;

--
-- Name: message_struct; Type: TYPE; Schema: public; Owner: i9ine
--

CREATE TYPE public.message_struct AS (
	id_ uuid,
	che_type text,
	content_ json,
	delivery_status text,
	created_at bigint,
	sender json,
	reply_target_msg json,
	cursor_ bigint
);


ALTER TYPE public.message_struct OWNER TO i9ine;

--
-- Name: msg_reaction_struct; Type: TYPE; Schema: public; Owner: i9ine
--

CREATE TYPE public.msg_reaction_struct AS (
	che_id uuid,
	che_type text,
	emoji text,
	reactor json,
	cursor_ bigint,
	to_msg json
);


ALTER TYPE public.msg_reaction_struct OWNER TO i9ine;

--
-- Name: new_comment_struct; Type: TYPE; Schema: public; Owner: i9ine
--

CREATE TYPE public.new_comment_struct AS (
	comment_id uuid,
	owner_user json,
	attachment_url text,
	comment_text text,
	at_ bigint,
	reactions_count integer,
	comments_count integer,
	me_reaction text,
	cursor_ bigint,
	ment_notif_ids uuid[],
	comment_notif_id uuid
);


ALTER TYPE public.new_comment_struct OWNER TO i9ine;

--
-- Name: new_post_struct; Type: TYPE; Schema: public; Owner: i9ine
--

CREATE TYPE public.new_post_struct AS (
	id_ uuid,
	type_ text,
	owner_user json,
	reposter_username text,
	media_urls text[],
	description text,
	created_at bigint,
	reactions_count integer,
	comments_count integer,
	reposts_count integer,
	saves_count integer,
	me_reaction text,
	me_saved boolean,
	me_reposted boolean,
	cursor_ bigint,
	ment_notif_ids uuid[]
);


ALTER TYPE public.new_post_struct OWNER TO i9ine;

--
-- Name: notif_struct; Type: TYPE; Schema: public; Owner: i9ine
--

CREATE TYPE public.notif_struct AS (
	id_ uuid,
	type_ text,
	at_ bigint,
	details json,
	unread boolean,
	cursor_ bigint,
	owner_username text
);


ALTER TYPE public.notif_struct OWNER TO i9ine;

--
-- Name: post_struct; Type: TYPE; Schema: public; Owner: i9ine
--

CREATE TYPE public.post_struct AS (
	id_ uuid,
	type_ text,
	owner_user json,
	reposter_username text,
	media_urls text[],
	description text,
	created_at bigint,
	reactions_count integer,
	comments_count integer,
	reposts_count integer,
	saves_count integer,
	me_reaction text,
	me_saved boolean,
	me_reposted boolean,
	cursor_ bigint
);


ALTER TYPE public.post_struct OWNER TO i9ine;

--
-- Name: user_profile_struct; Type: TYPE; Schema: public; Owner: i9ine
--

CREATE TYPE public.user_profile_struct AS (
	username text,
	name_ text,
	profile_pic_url text,
	bio text,
	posts_count integer,
	followers_count integer,
	followings_count integer,
	me_follow boolean,
	follows_me boolean
);


ALTER TYPE public.user_profile_struct OWNER TO i9ine;

--
-- Name: ack_msg(text, text, uuid[], text, bigint); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.ack_msg(from_user text, to_user text, msg_id_list uuid[], ack_val text, at_val bigint) RETURNS boolean
    LANGUAGE plpgsql
    AS $$BEGIN

IF ack_val = 'delivered' THEN
  UPDATE chat_history_entry che
  SET delivery_status = ack_val, delivered_at = at_val
  WHERE id_ = ANY(msg_id_list) AND type_ = 'message' AND delivery_status = 'sent' AND (SELECT EXISTS (SELECT 1 FROM chat_history_entry_in_chat WHERE owner_user = from_user AND partner_user = to_user AND che_id = che.id_));
  
  IF FOUND THEN
    UPDATE chats SET cursor_ = at_val
	WHERE owner_user = from_user AND partner_user = to_user;
    RETURN true;
  END IF;

  RETURN false;
ELSIF ack_val = 'read' THEN
	UPDATE chat_history_entry che
	SET delivery_status = ack_val, read_at = at_val, delivered_at = CASE WHEN delivered_at IS NULL THEN at_val ELSE delivered_at END
	WHERE id_ = ANY(msg_id_list) AND type_ = 'message' AND delivery_status <> 'read' AND (SELECT EXISTS (SELECT 1 FROM chat_history_entry_in_chat WHERE owner_user = from_user AND partner_user = to_user AND che_id = che.id_));
	
	IF NOT FOUND THEN
	  RETURN false;
	END IF;
	
	RETURN true;
END IF;

RAISE EXCEPTION
		USING
				ERRCODE = 'UX001',
				MESSAGE = 'invalid acknowledgment value';
END;
$$;


ALTER FUNCTION public.ack_msg(from_user text, to_user text, msg_id_list uuid[], ack_val text, at_val bigint) OWNER TO i9ine;

--
-- Name: comment_on_comment(text, uuid, text, text, bigint, text[]); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.comment_on_comment(commenter_username text, parent_comment_id_ uuid, comment_text_ text, attachment_url_ text, at__ bigint, mentions text[]) RETURNS public.new_comment_struct
    LANGUAGE plpgsql
    AS $_$
DECLARE
  res_comment new_comment_struct;
  ment_username text;
  ment_notif_ids uuid[];
  mnid uuid;
  parent_comment_owner_username text;
  comment_notif_id uuid;
BEGIN
  SELECT username FROM public.comments WHERE comment_id = parent_comment_id_ INTO parent_comment_owner_username;
  IF NOT FOUND THEN
    RETURN NULL;
  END IF;
  
  INSERT INTO public.comments(username, parent_comment_id, comment_text, attachment_url, at_)
  VALUES ($1, $2, $3, $4, $5)
  RETURNING
    comment_id,
	(SELECT json_build_object('username', username, 'name', name_, 'profile_pic_url', profile_pic_url) FROM users WHERE username = commenter_username),
	attachment_url_,
	comment_text_,
	at_, 0, 0, '', cursor_, null, null
  INTO res_comment;

  IF parent_comment_owner_username <> commenter_username THEN
    INSERT INTO notifications (notif_key, type_, owner_user, at_, details)
    VALUES (concat('user_',commenter_username,'_comment_on_comment_',parent_comment_id_), 'comment_on_comment', parent_comment_owner_username, (EXTRACT(EPOCH FROM now()) * 1000)::bigint, jsonb_build_object('on_comment_id', parent_comment_id_, 'commenter_username', commenter_username, 'comment_id', res_comment.comment_id))
    RETURNING id_ INTO comment_notif_id;
  END IF;

  FOREACH ment_username IN ARRAY mentions LOOP
    INSERT INTO comment_mentions (comment_id, username)
	VALUES (res_comment.comment_id, ment_username)
	ON CONFLICT ON CONSTRAINT no_dup_comment_ment DO NOTHING;

	IF FOUND AND ment_username <> commenter_username THEN
	  INSERT INTO notifications (notif_key, type_, owner_user, at_, details)
	  VALUES (concat('user_',ment_username,'_mentioned_in_comment_',res_comment.comment_id), 'mention_in_comment', ment_username, (EXTRACT(EPOCH FROM now()) * 1000)::bigint, jsonb_build_object('in_comment_id', res_comment.comment_id, 'mentioning_username', commenter_username))
	  RETURNING id_ INTO mnid;
	  
	  ment_notif_ids := array_append(ment_notif_ids, mnid);
	END IF;
  END LOOP;

  res_comment.ment_notif_ids = ment_notif_ids;
  res_comment.comment_notif_id = comment_notif_id;

  RETURN res_comment;
END;
$_$;


ALTER FUNCTION public.comment_on_comment(commenter_username text, parent_comment_id_ uuid, comment_text_ text, attachment_url_ text, at__ bigint, mentions text[]) OWNER TO i9ine;

--
-- Name: comment_on_post(text, uuid, text, text, bigint, text[]); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.comment_on_post(commenter_username text, post_id_ uuid, comment_text_ text, attachment_url_ text, at__ bigint, mentions text[]) RETURNS public.new_comment_struct
    LANGUAGE plpgsql
    AS $_$
DECLARE
  res_comment new_comment_struct;
  ment_username text;
  ment_notif_ids uuid[];
  mnid uuid;
  post_owner_username text;
  comment_notif_id uuid;
BEGIN
  SELECT owner_user FROM posts WHERE id_ = post_id_ INTO post_owner_username;
  IF NOT FOUND THEN
    RETURN NULL;
  END IF;
  
  INSERT INTO public.comments(username, post_id, comment_text, attachment_url, at_)
  VALUES ($1, $2, $3, $4, $5)
  RETURNING
    comment_id,
	(SELECT json_build_object('username', username, 'name', name_, 'profile_pic_url', profile_pic_url) FROM users WHERE username = commenter_username),
	attachment_url_,
	comment_text_,
	at_, 0, 0, '', cursor_, null, null
  INTO res_comment;

  IF post_owner_username <> commenter_username THEN
    INSERT INTO notifications (notif_key, type_, owner_user, at_, details)
    VALUES (concat('user_',commenter_username,'_comment_on_post_',post_id_), 'comment_on_post', post_owner_username, (EXTRACT(EPOCH FROM now()) * 1000)::bigint, jsonb_build_object('on_post_id', post_id_, 'commenter_username', commenter_username, 'comment_id', res_comment.comment_id))
    RETURNING id_ INTO comment_notif_id;
  END IF;

  FOREACH ment_username IN ARRAY mentions LOOP
    INSERT INTO comment_mentions (comment_id, username)
	VALUES (res_comment.comment_id, ment_username)
	ON CONFLICT ON CONSTRAINT no_dup_comment_ment DO NOTHING;

	IF FOUND AND ment_username <> commenter_username THEN
	  INSERT INTO notifications (notif_key, type_, owner_user, at_, details)
	  VALUES (concat('user_',ment_username,'_mentioned_in_comment_',res_comment.comment_id), 'mention_in_comment', ment_username, (EXTRACT(EPOCH FROM now()) * 1000)::bigint, jsonb_build_object('in_comment_id', res_comment.comment_id, 'mentioning_username', commenter_username))
	  RETURNING id_ INTO mnid;
	  
	  ment_notif_ids := array_append(ment_notif_ids, mnid);
	END IF;
  END LOOP;

  res_comment.ment_notif_ids = ment_notif_ids;
  res_comment.comment_notif_id = comment_notif_id;

  RETURN res_comment;
END;
$_$;


ALTER FUNCTION public.comment_on_post(commenter_username text, post_id_ uuid, comment_text_ text, attachment_url_ text, at__ bigint, mentions text[]) OWNER TO i9ine;

--
-- Name: delete_msg(text, text, uuid, text, bigint); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.delete_msg(from_user text, to_user text, msg_id uuid, deletefor text, at_val bigint) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN

IF deletefor = 'everyone' THEN
	UPDATE chat_history_entry_in_chat
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
	UPDATE chat_history_entry_in_chat
	SET deleted = true, deleted_at = at_val
	WHERE che_id = msg_id AND owner_user = from_user AND partner_user = to_user AND receipt IN ('sent', 'received');
	IF FOUND THEN
		RETURN true;
	END IF;
END IF;

RETURN false;
END;
$$;


ALTER FUNCTION public.delete_msg(from_user text, to_user text, msg_id uuid, deletefor text, at_val bigint) OWNER TO i9ine;

--
-- Name: fetch_chat_history(text, text, integer, integer, text); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.fetch_chat_history(owner_username text, partner_username text, in_limit integer, in_cursor integer, rel text) RETURNS SETOF public.che_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT 
	  /* general */
	  che.type_,
      che.id_,

	  /* type: message */
	  che.content_,
	  che.delivery_status,
	  che.created_at,
	  che.delivered_at,
	  che.read_at,
	  CASE WHEN sndr.username IS NULL THEN NULL ELSE json_build_object('username', sndr.username, 'profile_pic_url', sndr.profile_pic_url) END AS sender,
	  /* if  one is available */
	  (SELECT json_build_object('id', id_, 'sender_username', sender_username, 'content', content_) FROM chat_history_entry WHERE id_ = che.reply_to) AS reply_target_msg,

	  /* type: reaction */
	  CASE WHEN rctr.username IS NULL THEN NULL ELSE json_build_object('username', rctr.username, 'profile_pic_url', rctr.profile_pic_url) END AS reactor,
	  che.emoji,
	  (SELECT json_build_object('id', id_, 'sender_username', sender_username, 'content', content_) FROM chat_history_entry WHERE id_ = che.reaction_to) AS rxn_to_msg,

      /* general */
	  che.cursor_
    FROM chat_history_entry che
	INNER JOIN chat_history_entry_in_chat cheic ON cheic.che_id = che.id_ AND cheic.owner_user = owner_username AND cheic.partner_user = partner_username
	LEFT JOIN users sndr ON sndr.username = che.sender_username
	LEFT JOIN users rctr ON rctr.username = che.reactor_username
	WHERE in_cursor = 0 OR ((rel = 'after' AND che.cursor_ > in_cursor) OR (rel = 'before' AND che.cursor_ < in_cursor))
	ORDER BY che.cursor_ ASC
	LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.fetch_chat_history(owner_username text, partner_username text, in_limit integer, in_cursor integer, rel text) OWNER TO i9ine;

--
-- Name: fetch_notifs(uuid[]); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.fetch_notifs(notif_ids uuid[]) RETURNS SETOF public.notif_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT id_, 
    type_,
    at_,
	CASE type_ 
	  WHEN 'user_follow' THEN
	    json_build_object('follower_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'follower_username'))
	  WHEN 'mention_in_post' THEN json_build_object(
		  'in_post', (SELECT json_build_object('id', id_, 'description', description) FROM posts WHERE id_ = CAST(notif.details ->> 'in_post_id' AS uuid)),
		  'mentioning_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'mentioning_username')
		)
	  WHEN 'mention_in_comment' THEN json_build_object(
		  'in_comment', (SELECT json_build_object('id', comment_id, 'comment_text', comment_text) FROM public.comments WHERE comment_id = CAST(notif.details ->> 'in_comment_id' AS uuid)),
		  'mentioning_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'mentioning_username')
		)
	  WHEN 'reaction_to_post' THEN json_build_object(
		  'to_post', (SELECT json_build_object('id', id_, 'description', description) FROM posts WHERE id_ = CAST(notif.details ->> 'to_post_id' AS uuid)),
		  'reactor_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'reactor_username'),
		  'emoji', notif.details ->> 'emoji'
		)
	  WHEN 'reaction_to_comment' THEN json_build_object(
		  'to_comment', (SELECT json_build_object('id', comment_id, 'comment_text', comment_text) FROM public.comments WHERE comment_id = CAST(notif.details ->> 'to_comment_id' AS uuid)),
		  'reactor_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'reactor_username'),
		  'emoji', notif.details ->> 'emoji'
		)
	  WHEN 'comment_on_post' THEN json_build_object(
		  'on_post_id', notif.details ->> 'on_post_id',
		  'commenter_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'commenter_username'),
		  'comment', (SELECT json_build_object('id', comment_id, 'comment_text', comment_text) FROM public.comments WHERE comment_id = CAST(notif.details ->> 'comment_id' AS uuid))
		)
	  WHEN 'comment_on_comment' THEN json_build_object(
		  'on_comment_id', notif.details ->> 'on_comment_id',
		  'commenter_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'commenter_username'),
		  'comment', (SELECT json_build_object('id', comment_id, 'comment_text', comment_text) FROM public.comments WHERE comment_id = CAST(notif.details ->> 'comment_id' AS uuid))
		)
	  WHEN 'repost' THEN json_build_object(
		  'reposted_post', (SELECT json_build_object('id', id_, 'description', description) FROM posts WHERE id_ = CAST(notif.details ->> 'reposted_post_id' AS uuid)),
		  'reposter_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'reposter_username'),
		  'repost_id', notif.details ->> 'repost_id'
		)
	  ELSE null
	END AS details,
    unread,
	cursor_,
	owner_user AS owner_username
  FROM notifications notif
  WHERE notif.id_ = ANY(notif_ids);
END;
$$;


ALTER FUNCTION public.fetch_notifs(notif_ids uuid[]) OWNER TO i9ine;

--
-- Name: follow_user(text, text, bigint); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.follow_user(follower_username_ text, following_username_ text, follow_at bigint) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
  notif_id uuid;
BEGIN
  INSERT INTO follows (follower_username, following_username, at_)
  VALUES (follower_username_, following_username_, follow_at)
  ON CONFLICT ON CONSTRAINT no_dup_follow DO NOTHING;
  
  IF FOUND THEN
    INSERT INTO notifications (notif_key, type_, owner_user, at_, details)
	VALUES (concat('user_',follower_username_,'_follows_user_',following_username_), 'user_follow', following_username_, (EXTRACT(EPOCH FROM now()) * 1000)::bigint, jsonb_build_object('follower_username', follower_username_))
	RETURNING id_ INTO notif_id;

	RETURN notif_id;
  END IF;

  RETURN NULL;
END;
$$;


ALTER FUNCTION public.follow_user(follower_username_ text, following_username_ text, follow_at bigint) OWNER TO i9ine;

--
-- Name: get_comment(text, uuid); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_comment(client_username text, comment_id_ uuid) RETURNS SETOF public.comment_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT c.comment_id,
	  json_build_object('username', ou.username, 'name', ou.name_, 'profile_pic_url', ou.profile_pic_url) AS owner_user,
	  attachment_url,
	  comment_text,
	  c.at_,
	  reactions_count,
	  comments_count,
	  cr.emoji AS me_reaction,
	  c.cursor_
	FROM public.comments c
	INNER JOIN users ou ON ou.username = c.username
	LEFT JOIN comment_reactions cr ON cr.comment_id = c.comment_id AND cr.username = client_username
	WHERE c.comment_id = comment_id_;
END;
$$;


ALTER FUNCTION public.get_comment(client_username text, comment_id_ uuid) OWNER TO i9ine;

--
-- Name: get_comment_comments(text, uuid, integer, integer); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_comment_comments(client_username text, parent_comment_id_ uuid, in_limit integer, in_cursor integer) RETURNS SETOF public.comment_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT c.comment_id,
	  json_build_object('username', ou.username, 'name', ou.name_, 'profile_pic_url', ou.profile_pic_url) AS owner_user,
	  attachment_url,
	  comment_text,
	  c.at_,
	  reactions_count,
	  comments_count,
	  cr.emoji AS me_reaction,
	  c.cursor_
	FROM public.comments c
	INNER JOIN users ou ON ou.username = c.username
	LEFT JOIN comment_reactions cr ON cr.comment_id = c.comment_id AND cr.username = client_username
	WHERE c.parent_comment_id = parent_comment_id_ AND (in_cursor = 0 OR c.cursor_ < in_cursor)
	ORDER BY c.cursor_ DESC
	LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.get_comment_comments(client_username text, parent_comment_id_ uuid, in_limit integer, in_cursor integer) OWNER TO i9ine;

--
-- Name: get_comment_reactors(uuid, integer, integer); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_comment_reactors(comment_id_ uuid, in_limit integer, in_cursor integer) RETURNS TABLE(username text, name_ text, profile_pic_url text, emoji text, cursor_ bigint)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT u.username,
	  u.name_,
	  u.profile_pic_url,
	  cr.emoji,
	  cr.cursor_
	FROM comment_reactions cr
	INNER JOIN users u ON u.username = cr.username
	WHERE cr.comment_id = comment_id_ AND (in_cursor = 0 OR cr.cursor_ < in_cursor)
	ORDER BY cr.cursor_ DESC
	LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.get_comment_reactors(comment_id_ uuid, in_limit integer, in_cursor integer) OWNER TO i9ine;

--
-- Name: get_mentioned_posts(text, integer, integer); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_mentioned_posts(client_username text, in_limit integer, in_cursor integer) RETURNS SETOF public.post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT p.id_,
	  p.type_,
	  json_build_object('username', ou.username, 'name', ou.name_, 'profile_pic_url', ou.profile_pic_url) AS owner_user,
	  p.reposter_username,
	  p.media_urls,
	  p.description,
	  p.created_at,
	  p.reactions_count,
	  p.comments_count,
	  p.reposts_count,
	  p.saves_count,
	  pr.emoji AS me_reaction,
	  ps.username IS NOT NULL AS me_saved,
	  rep.reposter_username IS NOT NULL AS me_reposted,
	  p.cursor_
	FROM posts p
	INNER JOIN users ou ON ou.username = p.owner_user
	INNER JOIN post_mentions pm ON pm.post_id = p.id_ AND pm.username = client_username
	LEFT JOIN post_reactions pr ON pr.post_id = p.id_ AND pr.username = client_username
	LEFT JOIN post_saves ps ON ps.post_id = p.id_ AND ps.username = client_username
	LEFT JOIN posts rep ON rep.reposted_post_id = p.id_ AND rep.reposter_username = client_username
	WHERE in_cursor = 0 OR p.cursor_ < in_cursor
	ORDER BY p.cursor_ DESC
	LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.get_mentioned_posts(client_username text, in_limit integer, in_cursor integer) OWNER TO i9ine;

--
-- Name: get_my_chats(text, integer, integer); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_my_chats(owner_username text, in_limit integer, in_cursor integer) RETURNS TABLE(partner_user json)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY 
    SELECT json_build_object('username', pu.username, 'profile_pic_url', pu.profile_pic_url, 'presence', pu.presence) AS partner_user
	FROM chats c
	INNER JOIN users pu ON pu.username = c.partner_user
	WHERE c.owner_user = owner_username AND (in_cursor = 0 OR c.cursor_ > in_cursor)
	ORDER BY c.cursor_ DESC
	LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.get_my_chats(owner_username text, in_limit integer, in_cursor integer) OWNER TO i9ine;

--
-- Name: get_my_notifs(text, integer, integer); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_my_notifs(client_username text, in_limit integer, in_cursor integer) RETURNS SETOF public.notif_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT id_, 
    type_,
    at_,
	CASE type_ 
	  WHEN 'user_follow' THEN
	    json_build_object('follower_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'follower_username'))
	  WHEN 'mention_in_post' THEN json_build_object(
		  'in_post', (SELECT json_build_object('id', id_, 'description', description) FROM posts WHERE id_ = CAST(notif.details ->> 'in_post_id' AS uuid)),
		  'mentioning_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'mentioning_username')
		)
	  WHEN 'mention_in_comment' THEN json_build_object(
		  'in_comment', (SELECT json_build_object('id', comment_id, 'comment_text', comment_text) FROM public.comments WHERE comment_id = CAST(notif.details ->> 'in_comment_id' AS uuid)),
		  'mentioning_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'mentioning_username')
		)
	  WHEN 'reaction_to_post' THEN json_build_object(
		  'to_post', (SELECT json_build_object('id', id_, 'description', description) FROM posts WHERE id_ = CAST(notif.details ->> 'to_post_id' AS uuid)),
		  'reactor_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'reactor_username'),
		  'emoji', notif.details ->> 'emoji'
		)
	  WHEN 'reaction_to_comment' THEN json_build_object(
		  'to_comment', (SELECT json_build_object('id', comment_id, 'comment_text', comment_text) FROM public.comments WHERE comment_id = CAST(notif.details ->> 'to_comment_id' AS uuid)),
		  'reactor_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'reactor_username'),
		  'emoji', notif.details ->> 'emoji'
		)
	  WHEN 'comment_on_post' THEN json_build_object(
		  'on_post_id', notif.details ->> 'on_post_id',
		  'commenter_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'commenter_username'),
		  'comment', (SELECT json_build_object('id', comment_id, 'comment_text', comment_text) FROM public.comments WHERE comment_id = CAST(notif.details ->> 'comment_id' AS uuid))
		)
	  WHEN 'comment_on_comment' THEN json_build_object(
		  'on_comment_id', notif.details ->> 'on_comment_id',
		  'commenter_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'commenter_username'),
		  'comment', (SELECT json_build_object('id', comment_id, 'comment_text', comment_text) FROM public.comments WHERE comment_id = CAST(notif.details ->> 'comment_id' AS uuid))
		)
	  WHEN 'repost' THEN json_build_object(
		  'reposted_post', (SELECT json_build_object('id', id_, 'description', description) FROM posts WHERE id_ = CAST(notif.details ->> 'reposted_post_id' AS uuid)),
		  'reposter_user', (SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) FROM users WHERE username = notif.details ->> 'reposter_username'),
		  'repost_id', notif.details ->> 'repost_id'
		)
	  ELSE null
	END AS details,
    unread,
	cursor_,
	owner_user AS owner_username
  FROM notifications notif
  WHERE notif.owner_user = client_usesrname AND (in_cursor = 0 OR notif.cursor_ < in_cursor)
  ORDER BY notif.cursor_ DESC
  LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.get_my_notifs(client_username text, in_limit integer, in_cursor integer) OWNER TO i9ine;

--
-- Name: get_post(text, uuid); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_post(client_username text, post_id_ uuid) RETURNS SETOF public.post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT p.id_,
	  p.type_,
	  json_build_object('username', ou.username, 'name', ou.name_, 'profile_pic_url', ou.profile_pic_url) AS owner_user,
	  p.reposter_username,
	  p.media_urls,
	  p.description,
	  p.created_at,
	  p.reactions_count,
	  p.comments_count,
	  p.reposts_count,
	  p.saves_count,
	  pr.emoji AS me_reaction,
	  ps.username IS NOT NULL AS me_saved,
	  rep.reposter_username IS NOT NULL AS me_reposted,
	  p.cursor_
	FROM posts p
	INNER JOIN users ou ON ou.username = p.owner_user
	LEFT JOIN post_saves ps ON ps.post_id = p.id_ AND ps.username = client_username
	LEFT JOIN post_reactions pr ON pr.post_id = p.id_ AND pr.username = client_username
	LEFT JOIN posts rep ON rep.reposted_post_id = p.id_ AND rep.reposter_username = client_username
	WHERE p.id_ = post_id_;
END;
$$;


ALTER FUNCTION public.get_post(client_username text, post_id_ uuid) OWNER TO i9ine;

--
-- Name: get_post_comments(text, uuid, integer, integer); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_post_comments(client_username text, post_id_ uuid, in_limit integer, in_cursor integer) RETURNS SETOF public.comment_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT c.comment_id,
	  json_build_object('username', ou.username, 'name', ou.name_, 'profile_pic_url', ou.profile_pic_url) AS owner_user,
	  attachment_url,
	  comment_text,
	  c.at_,
	  reactions_count,
	  comments_count,
	  cr.emoji AS me_reaction,
	  c.cursor_
	FROM public.comments c
	INNER JOIN users ou ON ou.username = c.username
	LEFT JOIN comment_reactions cr ON cr.comment_id = c.comment_id AND cr.username = client_username
	WHERE c.post_id = post_id_ AND (in_cursor = 0 OR c.cursor_ < in_cursor)
	ORDER BY c.cursor_ DESC
	LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.get_post_comments(client_username text, post_id_ uuid, in_limit integer, in_cursor integer) OWNER TO i9ine;

--
-- Name: get_post_reactors(uuid, integer, integer); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_post_reactors(post_id_ uuid, in_limit integer, in_cursor integer) RETURNS TABLE(username text, name_ text, profile_pic_url text, emoji text, cursor_ bigint)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT u.username,
	  u.name_,
	  u.profile_pic_url,
	  pr.emoji,
	  pr.cursor_
	FROM post_reactions pr
	INNER JOIN users u ON u.username = pr.username
	WHERE pr.post_id = post_id_ AND (in_cursor = 0 OR pr.cursor_ < in_cursor)
	ORDER BY pr.cursor_ DESC
	LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.get_post_reactors(post_id_ uuid, in_limit integer, in_cursor integer) OWNER TO i9ine;

--
-- Name: get_reacted_posts(text, integer, integer); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_reacted_posts(client_username text, in_limit integer, in_cursor integer) RETURNS SETOF public.post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT p.id_,
	  p.type_,
	  json_build_object('username', ou.username, 'name', ou.name_, 'profile_pic_url', ou.profile_pic_url) AS owner_user,
	  p.reposter_username,
	  p.media_urls,
	  p.description,
	  p.created_at,
	  p.reactions_count,
	  p.comments_count,
	  p.reposts_count,
	  p.saves_count,
	  pr.emoji AS me_reaction,
	  ps.username IS NOT NULL AS me_saved,
	  rep.reposter_username IS NOT NULL AS me_reposted,
	  p.cursor_
	FROM posts p
	INNER JOIN users ou ON ou.username = p.owner_user
	INNER JOIN post_reactions pr ON pr.post_id = p.id_ AND pr.username = client_username
	LEFT JOIN post_saves ps ON ps.post_id = p.id_ AND ps.username = client_username
	LEFT JOIN posts rep ON rep.reposted_post_id = p.id_ AND rep.reposter_username = client_username
	WHERE in_cursor = 0 OR p.cursor_ < in_cursor
	ORDER BY p.cursor_ DESC
	LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.get_reacted_posts(client_username text, in_limit integer, in_cursor integer) OWNER TO i9ine;

--
-- Name: get_saved_posts(text, integer, integer); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_saved_posts(client_username text, in_limit integer, in_cursor integer) RETURNS SETOF public.post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT p.id_,
	  p.type_,
	  json_build_object('username', ou.username, 'name', ou.name_, 'profile_pic_url', ou.profile_pic_url) AS owner_user,
	  p.reposter_username,
	  p.media_urls,
	  p.description,
	  p.created_at,
	  p.reactions_count,
	  p.comments_count,
	  p.reposts_count,
	  p.saves_count,
	  pr.emoji AS me_reaction,
	  ps.username IS NOT NULL AS me_saved,
	  rep.reposter_username IS NOT NULL AS me_reposted,
	  p.cursor_
	FROM posts p
	INNER JOIN users ou ON ou.username = p.owner_user
	INNER JOIN post_saves ps ON ps.post_id = p.id_ AND ps.username = client_username
	LEFT JOIN post_reactions pr ON pr.post_id = p.id_ AND pr.username = client_username
	LEFT JOIN posts rep ON rep.reposted_post_id = p.id_ AND rep.reposter_username = client_username
	WHERE in_cursor = 0 OR p.cursor_ < in_cursor
	ORDER BY p.cursor_ DESC
	LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.get_saved_posts(client_username text, in_limit integer, in_cursor integer) OWNER TO i9ine;

--
-- Name: get_user_followers(text, text, integer, integer); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_user_followers(client_username text, target_username text, in_limit integer, in_cursor integer) RETURNS TABLE(username text, name_ text, profile_pic_url text, bio text, me_follow boolean, follows_me boolean, cursor_ bigint)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT u.username,
	  u.name_,
	  u.profile_pic_url,
	  u.bio,
	  flwer.follower_username IS NOT NULL AS me_follow,
	  flwng.following_username IS NOT NULL AS follows_me,
	  f.cursor_
	FROM follows f
	INNER JOIN users u ON u.username = f.follower_username
	LEFT JOIN follows flwer ON flwer.follower_username = client_username AND flwer.following_username = u.username
	LEFT JOIN follows flwng ON flwng.follower_username = u.username AND flwng.following_username = client_username
	WHERE f.following_username = target_username AND (in_cursor = 0 OR f.cursor_ < in_cursor)
	ORDER BY f.cursor_ DESC
	LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.get_user_followers(client_username text, target_username text, in_limit integer, in_cursor integer) OWNER TO i9ine;

--
-- Name: get_user_followings(text, text, integer, integer); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_user_followings(client_username text, target_username text, in_limit integer, in_cursor integer) RETURNS TABLE(username text, name_ text, profile_pic_url text, bio text, me_follow boolean, follows_me boolean, cursor_ bigint)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT u.username,
	  u.name_,
	  u.profile_pic_url,
	  u.bio,
	  flwer.follower_username IS NOT NULL AS me_follow,
	  flwng.following_username IS NOT NULL AS follows_me,
	  f.cursor_
	FROM follows f
	INNER JOIN users u ON u.username = f.following_username
	LEFT JOIN follows flwer ON flwer.follower_username = client_username AND flwer.following_username = u.username
	LEFT JOIN follows flwng ON flwng.follower_username = u.username AND flwng.following_username = client_username
	WHERE f.follower_username = target_username AND (in_cursor = 0 OR f.cursor_ < in_cursor)
	ORDER BY f.cursor_ DESC
	LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.get_user_followings(client_username text, target_username text, in_limit integer, in_cursor integer) OWNER TO i9ine;

--
-- Name: get_user_posts(text, text, integer, integer); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_user_posts(client_username text, target_username text, in_limit integer, in_cursor integer) RETURNS SETOF public.post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT p.id_,
	  p.type_,
	  json_build_object('username', ou.username, 'name', ou.name_, 'profile_pic_url', ou.profile_pic_url) AS owner_user,
	  p.reposter_username,
	  p.media_urls,
	  p.description,
	  p.created_at,
	  p.reactions_count,
	  p.comments_count,
	  p.reposts_count,
	  p.saves_count,
	  pr.emoji AS me_reaction,
	  ps.username IS NOT NULL AS me_saved,
	  rep.reposter_username IS NOT NULL AS me_reposted,
	  p.cursor_
	FROM posts p
	INNER JOIN users ou ON ou.username = p.owner_user
	LEFT JOIN post_saves ps ON ps.post_id = p.id_ AND ps.username = client_username
	LEFT JOIN post_reactions pr ON pr.post_id = p.id_ AND pr.username = client_username
	LEFT JOIN posts rep ON rep.reposted_post_id = p.id_ AND rep.reposter_username = client_username
	WHERE p.owner_user = target_username AND (in_cursor = 0 OR p.cursor_ < in_cursor)
	ORDER BY p.cursor_ DESC
	LIMIT in_limit;
END;
$$;


ALTER FUNCTION public.get_user_posts(client_username text, target_username text, in_limit integer, in_cursor integer) OWNER TO i9ine;

--
-- Name: get_user_profile(text, text); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.get_user_profile(client_username text, target_username text) RETURNS SETOF public.user_profile_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
    SELECT username,
	  name_,
	  profile_pic_url,
	  bio,
	  posts_count,
	  followers_count,
	  followings_count,
	  flwer.follower_username IS NOT NULL AS me_follow,
	  flwng.following_username IS NOT NULL AS follows_me
	FROM users u
	LEFT JOIN follows flwer ON flwer.follower_username = client_username AND flwer.following_username = u.username
	LEFT JOIN follows flwng ON flwng.follower_username = u.username AND flwng.following_username = client_username
	WHERE u.username = target_username;
END;
$$;


ALTER FUNCTION public.get_user_profile(client_username text, target_username text) OWNER TO i9ine;

--
-- Name: new_post(text, text, text[], text, bigint, text[], text[]); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.new_post(owner_username text, type__ text, media_urls_ text[], description_ text, created_at_ bigint, mentions text[], hashtags text[]) RETURNS public.new_post_struct
    LANGUAGE plpgsql
    AS $$
DECLARE
  res_post new_post_struct;
  ment_user text;
  ht_name text;
  ment_notif_ids uuid[];
  mnid uuid;
BEGIN
  INSERT INTO posts (owner_user, type_, media_urls, description, created_at)
  VALUES (owner_username, type__, media_urls_, description_, created_at_)
  RETURNING
    id_,
	type_,
	(SELECT json_build_object('username', username, 'name', name_, 'profile_pic_url', profile_pic_url) FROM users WHERE username = owner_username),
	'',
	media_urls, description, created_at, 0, 0, 0, 0, '', false, false, cursor_, null 
  INTO res_post;

  UPDATE users
  SET posts_count = posts_count + 1
  WHERE username = owner_username;

  FOREACH ment_user IN ARRAY mentions LOOP
    INSERT INTO post_mentions (post_id, username)
	VALUES (res_post.id_, ment_user)
	ON CONFLICT ON CONSTRAINT no_dup_post_ment DO NOTHING;

	IF FOUND AND ment_user <> owner_username THEN
	  INSERT INTO notifications (notif_key, type_, owner_user, at_, details)
	  VALUES (concat('user_',ment_user,'_mentioned_in_post_',res_post.id_), 'mention_in_post', ment_user, (EXTRACT(EPOCH FROM now()) * 1000)::bigint, jsonb_build_object('in_post_id', res_post.id_, 'mentioning_username', owner_username))
	  RETURNING id_ INTO mnid;
	  
	  ment_notif_ids := array_append(ment_notif_ids, mnid);
	END IF;
  END LOOP;

  FOREACH ht_name IN ARRAY hashtags LOOP
    INSERT INTO hashtags (htname)
	VALUES (ht_name)
    ON CONFLICT ON CONSTRAINT hashtags_pkey DO NOTHING;

	INSERT INTO post_hashtags (post_id, htname)
	VALUES (res_post.id_, ht_name)
	ON CONFLICT ON CONSTRAINT no_dup_htname DO NOTHING;
  END LOOP;

  res_post.ment_notif_ids = ment_notif_ids;

  RETURN res_post;
END;
$$;


ALTER FUNCTION public.new_post(owner_username text, type__ text, media_urls_ text[], description_ text, created_at_ bigint, mentions text[], hashtags text[]) OWNER TO i9ine;

--
-- Name: react_to_comment(text, uuid, text, bigint); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.react_to_comment(reactor_username text, comment_id_ uuid, emoji_ text, r_at bigint) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
  comment_owner_username text;
  notif_id uuid;
BEGIN
  SELECT username FROM public.comments WHERE comment_id = comment_id_ INTO comment_owner_username;
  
  INSERT INTO comment_reactions(username, comment_id, emoji, at_)
  VALUES (reactor_username, comment_id_, emoji_, r_at)
  ON CONFLICT ON CONSTRAINT no_dup_comment_rxn DO UPDATE 
  SET emoji = emoji_, at_ = r_at;

  IF FOUND AND comment_owner_username <> reactor_username THEN
    INSERT INTO notifications (notif_key, type_, owner_user, at_, details)
	VALUES (concat('user_',reactor_username,'_reaction_to_comment_',comment_id_), 'reaction_to_comment', comment_owner_username, (EXTRACT(EPOCH FROM now()) * 1000)::bigint, jsonb_build_object('to_comment_id', comment_id_, 'reactor_username', reactor_username, 'emoji', emoji_))
	ON CONFLICT ON CONSTRAINT unique_notif_key DO UPDATE
	SET details = EXCLUDED.details
	RETURNING id_ INTO notif_id;

	RETURN notif_id;
  END IF;

  RETURN NULL;
END;
$$;


ALTER FUNCTION public.react_to_comment(reactor_username text, comment_id_ uuid, emoji_ text, r_at bigint) OWNER TO i9ine;

--
-- Name: react_to_msg(text, text, uuid, text, bigint); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.react_to_msg(from_user text, to_user text, msg_id uuid, emoji_val text, at_val bigint) RETURNS public.msg_reaction_struct
    LANGUAGE plpgsql
    AS $$
DECLARE
msg_in_chat bool;
che_id_val uuid;
cursor_val bigint;
reactor_user json;
rxn_to_msg json;
BEGIN

SELECT EXISTS (SELECT 1 FROM chat_history_entry_in_chat 
WHERE owner_user = from_user AND partner_user = to_user AND che_id = msg_id AND (
SELECT type_ FROM chat_history_entry WHERE id_ = msg_id) = 'message'
)
INTO msg_in_chat;

IF NOT msg_in_chat THEN
RAISE EXCEPTION
	USING
			ERRCODE = 'UX001',
			MESSAGE = 'you do not have a chat with the specified user or the specified message does not exist in the chat';
END IF;

INSERT INTO chat_history_entry (type_, reactor_username, emoji, reaction_at, reaction_to)
VALUES ('reaction', from_user, emoji_val, at_val, msg_id)
ON CONFLICT ON CONSTRAINT no_dup_msg_rxn DO UPDATE 
SET emoji = emoji_val, reaction_at = at_val
RETURNING id_, cursor_ INTO che_id_val, cursor_val;

INSERT INTO chat_history_entry_in_chat (owner_user, partner_user, che_id, receipt)
VALUES (from_user, to_user, che_id_val, 'sent')
ON CONFLICT ON CONSTRAINT no_dup_che DO NOTHING;

INSERT INTO chat_history_entry_in_chat (owner_user, partner_user, che_id, receipt)
VALUES (to_user, from_user, che_id_val, 'received')
ON CONFLICT ON CONSTRAINT no_dup_che DO NOTHING;

SELECT json_build_object('username', username, 'name', name_, 'profile_pic_url', profile_pic_url) INTO reactor_user
FROM users WHERE username = from_user;

SELECT json_build_object('id', id_, 'sender_username', sender_username, 'content', content_) INTO rxn_to_msg
FROM chat_history_entry WHERE id_ = msg_id;

RETURN ROW(che_id_val, 'reaction', emoji_val, reactor_user, cursor_val, rxn_to_msg)::msg_reaction_struct;

END;
$$;


ALTER FUNCTION public.react_to_msg(from_user text, to_user text, msg_id uuid, emoji_val text, at_val bigint) OWNER TO i9ine;

--
-- Name: react_to_post(text, uuid, text, bigint); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.react_to_post(reactor_username text, post_id_ uuid, emoji_ text, r_at bigint) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
  post_owner_username text;
  notif_id uuid;
BEGIN
  SELECT owner_user FROM posts WHERE id_ = post_id_ INTO post_owner_username;
  
  INSERT INTO post_reactions(username, post_id, emoji, at_)
  VALUES (reactor_username, post_id_, emoji_, r_at)
  ON CONFLICT ON CONSTRAINT no_dup_post_rxn DO UPDATE 
  SET emoji = emoji_, at_ = r_at;

  IF FOUND AND post_owner_username <> reactor_username THEN
    INSERT INTO notifications (notif_key, type_, owner_user, at_, details)
	VALUES (concat('user_',reactor_username,'_reaction_to_post_',post_id_), 'reaction_to_post', post_owner_username, (EXTRACT(EPOCH FROM now()) * 1000)::bigint, jsonb_build_object('to_post_id', post_id_, 'reactor_username', reactor_username, 'emoji', emoji_))
	ON CONFLICT ON CONSTRAINT unique_notif_key DO UPDATE
	SET details = EXCLUDED.details
	RETURNING id_ INTO notif_id;

	RETURN notif_id;
  END IF;

  RETURN NULL;
END;
$$;


ALTER FUNCTION public.react_to_post(reactor_username text, post_id_ uuid, emoji_ text, r_at bigint) OWNER TO i9ine;

--
-- Name: remove_msg_reaction(text, text, uuid); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.remove_msg_reaction(from_user text, to_user text, msg_id uuid) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE 
msg_in_chat bool;
che_id_val uuid;
BEGIN
SELECT EXISTS (SELECT 1 FROM chat_history_entry_in_chat 
WHERE owner_user = from_user AND partner_user = to_user AND che_id = msg_id AND (
	SELECT type_ FROM chat_history_entry WHERE id_ = msg_id) = 'message'
)
INTO msg_in_chat;

IF NOT msg_in_chat THEN
	RAISE EXCEPTION
	USING
			ERRCODE = 'UX001',
			MESSAGE = 'you do not have a chat with the specified user or the specified message does not exist in the chat';
END IF;
			
DELETE FROM chat_history_entry
WHERE reactor_username = from_user AND reaction_to = msg_id
RETURNING id_ INTO che_id_val;

RETURN che_id_val;
END;
$$;


ALTER FUNCTION public.remove_msg_reaction(from_user text, to_user text, msg_id uuid) OWNER TO i9ine;

--
-- Name: reply_to_msg(text, text, json, bigint, uuid); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.reply_to_msg(from_user text, to_user text, content_val json, created_at_val bigint, reply_target_msg_id uuid) RETURNS public.message_struct
    LANGUAGE plpgsql
    AS $$DECLARE
che_id_val uuid;
cursor_val bigint;
reply_target_msg json;
sender_user json;
BEGIN

SELECT 1 FROM chat_history_entry_in_chat 
WHERE owner_user = from_user AND partner_user = to_user AND che_id = reply_target_msg_id AND (SELECT type_ FROM chat_history_entry WHERE id_ = reply_target_msg_id) = 'message';

IF NOT FOUND THEN
RAISE EXCEPTION
	USING
			ERRCODE = 'UX001',
			MESSAGE = 'you do not have a chat with the specified user or the specified message does not exist in the chat';
END IF;

INSERT INTO chats (owner_user, partner_user)
VALUES (from_user, to_user)
ON CONFLICT ON CONSTRAINT ucu_pkey DO UPDATE
SET cursor_ = created_at_val;

INSERT INTO chats (owner_user, partner_user)
VALUES (to_user, from_user)
ON CONFLICT ON CONSTRAINT ucu_pkey DO NOTHING;

INSERT INTO chat_history_entry (type_, content_, sender_username, delivery_status, created_at, reply_to)
VALUES ('message', content_val, from_user, 'sent', created_at_val, reply_target_msg_id)
RETURNING id_, cursor_ INTO che_id_val, cursor_val;

INSERT INTO chat_history_entry_in_chat (owner_user, partner_user, che_id, receipt)
VALUES (from_user, to_user, che_id_val, 'sent');

INSERT INTO chat_history_entry_in_chat (owner_user, partner_user, che_id, receipt)
VALUES (to_user, from_user, che_id_val, 'received');

SELECT json_build_object('username', username, 'name', name_, 'profile_pic_url', profile_pic_url) INTO sender_user
FROM users WHERE username = from_user;

SELECT json_build_object('id', id_, 'content', content_, 'sender_username', sender_username)
FROM chat_history_entry WHERE id_ = reply_target_msg_id
INTO reply_target_msg;

RETURN ROW(che_id_val, 'message', content_val, 'sent', created_at_val, sender_user, reply_target_msg, cursor_val)::message_struct;

END;
$$;


ALTER FUNCTION public.reply_to_msg(from_user text, to_user text, content_val json, created_at_val bigint, reply_target_msg_id uuid) OWNER TO i9ine;

--
-- Name: repost(text, uuid); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.repost(reposter_username_ text, post_id_ uuid, OUT repost_id uuid, OUT repost_cursor bigint, OUT repost_notif_id uuid) RETURNS record
    LANGUAGE plpgsql
    AS $_$
DECLARE
  post_owner_username text;
BEGIN
  SELECT owner_user FROM posts WHERE id_ = post_id_ INTO post_owner_username;
  IF NOT FOUND THEN
    RETURN;
  END IF;
  
  INSERT INTO posts (owner_user, type_, media_urls, description, created_at, reposter_username, reposted_post_id)
  SELECT owner_user, type_, media_urls, description, created_at, $1, $2 FROM posts WHERE id_ = $2
  RETURNING id_, cursor_ INTO repost_id, repost_cursor;

  IF reposter_username_ <> post_owner_username THEN
    INSERT INTO notifications (notif_key, type_, owner_user, at_, details)
    VALUES (concat('user_',reposter_username_,'_reposted_post_',post_id_), 'repost', post_owner_username, (EXTRACT(EPOCH FROM now()) * 1000)::bigint, jsonb_build_object('reposted_post_id', post_id_, 'reposter_username', reposter_username_, 'repost_id', repost_id))
    RETURNING id_ INTO repost_notif_id;
  END IF;

  RETURN;
END;
$_$;


ALTER FUNCTION public.repost(reposter_username_ text, post_id_ uuid, OUT repost_id uuid, OUT repost_cursor bigint, OUT repost_notif_id uuid) OWNER TO i9ine;

--
-- Name: send_message(text, text, json, bigint); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.send_message(from_user text, to_user text, content_val json, created_at_val bigint) RETURNS public.message_struct
    LANGUAGE plpgsql
    AS $$
DECLARE
che_id_val uuid;
cursor_val bigint;
sender_user json;
BEGIN


INSERT INTO chats (owner_user, partner_user)
VALUES (from_user, to_user)
ON CONFLICT ON CONSTRAINT ucu_pkey DO UPDATE
SET cursor_ = created_at_val;

INSERT INTO chats (owner_user, partner_user)
VALUES (to_user, from_user)
ON CONFLICT ON CONSTRAINT ucu_pkey DO NOTHING;

INSERT INTO chat_history_entry (type_, content_, sender_username, delivery_status, created_at)
VALUES ('message', content_val, from_user, 'sent', created_at_val)
RETURNING id_, cursor_ INTO che_id_val, cursor_val;

INSERT INTO chat_history_entry_in_chat (owner_user, partner_user, che_id, receipt)
VALUES (from_user, to_user, che_id_val, 'sent');

INSERT INTO chat_history_entry_in_chat (owner_user, partner_user, che_id, receipt)
VALUES (to_user, from_user, che_id_val, 'received');

SELECT json_build_object('username', username, 'name', name_, 'profile_pic_url', profile_pic_url) INTO sender_user
FROM users WHERE username = from_user;

RETURN ROW(che_id_val, 'message', content_val, 'sent', created_at_val, sender_user, null, cursor_val)::message_struct;
END;
$$;


ALTER FUNCTION public.send_message(from_user text, to_user text, content_val json, created_at_val bigint) OWNER TO i9ine;

--
-- Name: uncomment_on_comment(text, uuid, uuid); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.uncomment_on_comment(commenter_username text, parent_comment_id_ uuid, comment_id_ uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $_$
BEGIN
  DELETE FROM public.comments
  WHERE username = $1 AND parent_comment_id = $2 AND comment_id = $3;
  
  IF FOUND THEN
    DELETE FROM notifications
	WHERE notif_key = concat('user_',commenter_username,'_comment_on_comment_',parent_comment_id_);

	RETURN true;
  END IF;

  RETURN false;
END;
$_$;


ALTER FUNCTION public.uncomment_on_comment(commenter_username text, parent_comment_id_ uuid, comment_id_ uuid) OWNER TO i9ine;

--
-- Name: uncomment_on_post(text, uuid, uuid); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.uncomment_on_post(commenter_username text, post_id_ uuid, comment_id_ uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $_$
BEGIN
  DELETE FROM public.comments
  WHERE username = $1 AND post_id = $2 AND comment_id = $3;
  
  IF FOUND THEN
    DELETE FROM notifications
	WHERE notif_key = concat('user_',commenter_username,'_comment_on_post_',post_id_);

	RETURN true;
  END IF;

  RETURN false;
END;
$_$;


ALTER FUNCTION public.uncomment_on_post(commenter_username text, post_id_ uuid, comment_id_ uuid) OWNER TO i9ine;

--
-- Name: unfollow_user(text, text); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.unfollow_user(follower_username_ text, following_username_ text) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  DELETE FROM follows
  WHERE follower_username = follower_username_ AND following_username = following_username_;
  
  IF FOUND THEN
    DELETE FROM notifications
	WHERE notif_key = concat('user_',follower_username_,'_follows_user_',following_username_);

	RETURN true;
  END IF;

  RETURN false;
END;
$$;


ALTER FUNCTION public.unfollow_user(follower_username_ text, following_username_ text) OWNER TO i9ine;

--
-- Name: unreact_to_comment(text, uuid); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.unreact_to_comment(reactor_username text, comment_id_ uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $_$
BEGIN
  DELETE FROM comment_reactions
  WHERE username = $1 AND comment_id = $2;
  
  IF FOUND THEN
    DELETE FROM notifications
	WHERE notif_key = concat('user_',reactor_username,'_reaction_to_comment_',comment_id_);

	RETURN true;
  END IF;

  RETURN false;
END;
$_$;


ALTER FUNCTION public.unreact_to_comment(reactor_username text, comment_id_ uuid) OWNER TO i9ine;

--
-- Name: unreact_to_post(text, uuid); Type: FUNCTION; Schema: public; Owner: i9ine
--

CREATE FUNCTION public.unreact_to_post(reactor_username text, post_id_ uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  DELETE FROM post_reactions
  WHERE username = reactor_username AND post_id = post_id_;
  
  IF FOUND THEN
    DELETE FROM notifications
	WHERE notif_key = concat('user_',reactor_username,'_reaction_to_post_',post_id_);

	RETURN true;
  END IF;

  RETURN false;
END;
$$;


ALTER FUNCTION public.unreact_to_post(reactor_username text, post_id_ uuid) OWNER TO i9ine;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: chat_history_entry; Type: TABLE; Schema: public; Owner: i9ine
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
    cursor_ bigint NOT NULL,
    CONSTRAINT chat_history_entry_delivery_status_check CHECK ((delivery_status = ANY (ARRAY['sent'::text, 'delivered'::text, 'read'::text]))),
    CONSTRAINT chat_history_entry_type__check CHECK ((type_ = ANY (ARRAY['message'::text, 'reaction'::text])))
);


ALTER TABLE public.chat_history_entry OWNER TO i9ine;

--
-- Name: chat_history_entry_cursor__seq; Type: SEQUENCE; Schema: public; Owner: i9ine
--

CREATE SEQUENCE public.chat_history_entry_cursor__seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.chat_history_entry_cursor__seq OWNER TO i9ine;

--
-- Name: chat_history_entry_cursor__seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9ine
--

ALTER SEQUENCE public.chat_history_entry_cursor__seq OWNED BY public.chat_history_entry.cursor_;


--
-- Name: chat_history_entry_in_chat; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.chat_history_entry_in_chat (
    owner_user text NOT NULL,
    partner_user text NOT NULL,
    che_id uuid NOT NULL,
    receipt text NOT NULL,
    deleted boolean,
    deleted_at bigint,
    CONSTRAINT chat_history_entry_in_chat_receipt_check CHECK ((receipt = ANY (ARRAY['sent'::text, 'received'::text])))
);


ALTER TABLE public.chat_history_entry_in_chat OWNER TO i9ine;

--
-- Name: chats; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.chats (
    owner_user text NOT NULL,
    partner_user text NOT NULL,
    cursor_ bigint
);


ALTER TABLE public.chats OWNER TO i9ine;

--
-- Name: comment_mentions; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.comment_mentions (
    comment_id uuid NOT NULL,
    username text NOT NULL
);


ALTER TABLE public.comment_mentions OWNER TO i9ine;

--
-- Name: comment_reactions; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.comment_reactions (
    username text NOT NULL,
    comment_id uuid NOT NULL,
    emoji text NOT NULL,
    at_ bigint,
    cursor_ bigint NOT NULL
);


ALTER TABLE public.comment_reactions OWNER TO i9ine;

--
-- Name: comment_reactions_cursor__seq; Type: SEQUENCE; Schema: public; Owner: i9ine
--

CREATE SEQUENCE public.comment_reactions_cursor__seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.comment_reactions_cursor__seq OWNER TO i9ine;

--
-- Name: comment_reactions_cursor__seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9ine
--

ALTER SEQUENCE public.comment_reactions_cursor__seq OWNED BY public.comment_reactions.cursor_;


--
-- Name: comments; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.comments (
    comment_id uuid DEFAULT gen_random_uuid() NOT NULL,
    username text NOT NULL,
    parent_comment_id uuid,
    post_id uuid,
    comment_text text NOT NULL,
    attachment_url text NOT NULL,
    deleted boolean DEFAULT false,
    deleted_at bigint,
    at_ bigint,
    cursor_ bigint NOT NULL,
    reactions_count integer DEFAULT 0 CONSTRAINT comments_reaction_counts_not_null NOT NULL,
    comments_count integer DEFAULT 0 NOT NULL,
    CONSTRAINT on_post_xor_on_comment CHECK ((((post_id IS NULL) AND (parent_comment_id IS NOT NULL)) OR ((post_id IS NOT NULL) AND (parent_comment_id IS NULL))))
);


ALTER TABLE public.comments OWNER TO i9ine;

--
-- Name: comments_cursor__seq; Type: SEQUENCE; Schema: public; Owner: i9ine
--

CREATE SEQUENCE public.comments_cursor__seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.comments_cursor__seq OWNER TO i9ine;

--
-- Name: comments_cursor__seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9ine
--

ALTER SEQUENCE public.comments_cursor__seq OWNED BY public.comments.cursor_;


--
-- Name: follows; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.follows (
    follower_username text NOT NULL,
    following_username text NOT NULL,
    at_ bigint,
    cursor_ bigint NOT NULL
);


ALTER TABLE public.follows OWNER TO i9ine;

--
-- Name: follows_cursor__seq; Type: SEQUENCE; Schema: public; Owner: i9ine
--

CREATE SEQUENCE public.follows_cursor__seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.follows_cursor__seq OWNER TO i9ine;

--
-- Name: follows_cursor__seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9ine
--

ALTER SEQUENCE public.follows_cursor__seq OWNED BY public.follows.cursor_;


--
-- Name: hashtags; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.hashtags (
    htname text NOT NULL
);


ALTER TABLE public.hashtags OWNER TO i9ine;

--
-- Name: notifications; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.notifications (
    id_ uuid DEFAULT gen_random_uuid() NOT NULL,
    type_ text NOT NULL,
    owner_user text NOT NULL,
    at_ bigint,
    details jsonb NOT NULL,
    cursor_ bigint NOT NULL,
    unread boolean DEFAULT true NOT NULL,
    notif_key text NOT NULL,
    CONSTRAINT notifications_type__check CHECK ((type_ = ANY (ARRAY['user_follow'::text, 'mention_in_post'::text, 'mention_in_comment'::text, 'comment_on_post'::text, 'comment_on_comment'::text, 'reaction_to_post'::text, 'reaction_to_comment'::text, 'repost'::text])))
);


ALTER TABLE public.notifications OWNER TO i9ine;

--
-- Name: notifications_cursor__seq; Type: SEQUENCE; Schema: public; Owner: i9ine
--

CREATE SEQUENCE public.notifications_cursor__seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.notifications_cursor__seq OWNER TO i9ine;

--
-- Name: notifications_cursor__seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9ine
--

ALTER SEQUENCE public.notifications_cursor__seq OWNED BY public.notifications.cursor_;


--
-- Name: post_hashtags; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.post_hashtags (
    post_id uuid NOT NULL,
    htname text NOT NULL
);


ALTER TABLE public.post_hashtags OWNER TO i9ine;

--
-- Name: post_mentions; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.post_mentions (
    post_id uuid NOT NULL,
    username text NOT NULL
);


ALTER TABLE public.post_mentions OWNER TO i9ine;

--
-- Name: post_reactions; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.post_reactions (
    username text NOT NULL,
    post_id uuid NOT NULL,
    emoji text NOT NULL,
    at_ bigint,
    cursor_ bigint NOT NULL
);


ALTER TABLE public.post_reactions OWNER TO i9ine;

--
-- Name: post_reactions_cursor__seq; Type: SEQUENCE; Schema: public; Owner: i9ine
--

CREATE SEQUENCE public.post_reactions_cursor__seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.post_reactions_cursor__seq OWNER TO i9ine;

--
-- Name: post_reactions_cursor__seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9ine
--

ALTER SEQUENCE public.post_reactions_cursor__seq OWNED BY public.post_reactions.cursor_;


--
-- Name: post_saves; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.post_saves (
    username text NOT NULL,
    post_id uuid NOT NULL,
    cursor_ bigint NOT NULL
);


ALTER TABLE public.post_saves OWNER TO i9ine;

--
-- Name: post_saves_cursor__seq; Type: SEQUENCE; Schema: public; Owner: i9ine
--

CREATE SEQUENCE public.post_saves_cursor__seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.post_saves_cursor__seq OWNER TO i9ine;

--
-- Name: post_saves_cursor__seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9ine
--

ALTER SEQUENCE public.post_saves_cursor__seq OWNED BY public.post_saves.cursor_;


--
-- Name: posts; Type: TABLE; Schema: public; Owner: i9ine
--

CREATE TABLE public.posts (
    id_ uuid DEFAULT gen_random_uuid() NOT NULL,
    owner_user text NOT NULL,
    type_ text NOT NULL,
    media_urls text[] NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    deleted boolean DEFAULT false,
    reposter_username text,
    reposted_post_id uuid,
    created_at bigint,
    deleted_at bigint,
    cursor_ bigint NOT NULL,
    reactions_count integer DEFAULT 0 NOT NULL,
    comments_count integer DEFAULT 0 NOT NULL,
    reposts_count integer DEFAULT 0 NOT NULL,
    saves_count integer DEFAULT 0 NOT NULL,
    CONSTRAINT posts_type__check CHECK ((type_ = ANY (ARRAY['photo:portrait'::text, 'photo:square'::text, 'photo:landscape'::text, 'video:portrait'::text, 'video:square'::text, 'video:landscape'::text, 'reel'::text])))
);


ALTER TABLE public.posts OWNER TO i9ine;

--
-- Name: posts_cursor__seq; Type: SEQUENCE; Schema: public; Owner: i9ine
--

CREATE SEQUENCE public.posts_cursor__seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.posts_cursor__seq OWNER TO i9ine;

--
-- Name: posts_cursor__seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9ine
--

ALTER SEQUENCE public.posts_cursor__seq OWNED BY public.posts.cursor_;


--
-- Name: users; Type: TABLE; Schema: public; Owner: i9ine
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
    followers_count integer DEFAULT 0 NOT NULL,
    followings_count integer DEFAULT 0 NOT NULL,
    posts_count integer DEFAULT 0 NOT NULL,
    CONSTRAINT users_presence_check CHECK ((presence = ANY (ARRAY['online'::text, 'offline'::text])))
);


ALTER TABLE public.users OWNER TO i9ine;

--
-- Name: chat_history_entry cursor_; Type: DEFAULT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chat_history_entry ALTER COLUMN cursor_ SET DEFAULT nextval('public.chat_history_entry_cursor__seq'::regclass);


--
-- Name: comment_reactions cursor_; Type: DEFAULT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.comment_reactions ALTER COLUMN cursor_ SET DEFAULT nextval('public.comment_reactions_cursor__seq'::regclass);


--
-- Name: comments cursor_; Type: DEFAULT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.comments ALTER COLUMN cursor_ SET DEFAULT nextval('public.comments_cursor__seq'::regclass);


--
-- Name: follows cursor_; Type: DEFAULT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.follows ALTER COLUMN cursor_ SET DEFAULT nextval('public.follows_cursor__seq'::regclass);


--
-- Name: notifications cursor_; Type: DEFAULT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.notifications ALTER COLUMN cursor_ SET DEFAULT nextval('public.notifications_cursor__seq'::regclass);


--
-- Name: post_reactions cursor_; Type: DEFAULT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_reactions ALTER COLUMN cursor_ SET DEFAULT nextval('public.post_reactions_cursor__seq'::regclass);


--
-- Name: post_saves cursor_; Type: DEFAULT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_saves ALTER COLUMN cursor_ SET DEFAULT nextval('public.post_saves_cursor__seq'::regclass);


--
-- Name: posts cursor_; Type: DEFAULT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.posts ALTER COLUMN cursor_ SET DEFAULT nextval('public.posts_cursor__seq'::regclass);


--
-- Name: chat_history_entry chat_history_entry_pkey; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chat_history_entry
    ADD CONSTRAINT chat_history_entry_pkey PRIMARY KEY (id_);


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (comment_id);


--
-- Name: hashtags hashtags_pkey; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.hashtags
    ADD CONSTRAINT hashtags_pkey PRIMARY KEY (htname);


--
-- Name: chat_history_entry_in_chat no_dup_che; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chat_history_entry_in_chat
    ADD CONSTRAINT no_dup_che UNIQUE (owner_user, partner_user, che_id);


--
-- Name: comment_mentions no_dup_comment_ment; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.comment_mentions
    ADD CONSTRAINT no_dup_comment_ment UNIQUE (comment_id, username);


--
-- Name: comment_reactions no_dup_comment_rxn; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.comment_reactions
    ADD CONSTRAINT no_dup_comment_rxn UNIQUE (username, comment_id);


--
-- Name: follows no_dup_follow; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT no_dup_follow UNIQUE (follower_username, following_username);


--
-- Name: post_hashtags no_dup_htname; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_hashtags
    ADD CONSTRAINT no_dup_htname UNIQUE (post_id, htname);


--
-- Name: chat_history_entry no_dup_msg_rxn; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chat_history_entry
    ADD CONSTRAINT no_dup_msg_rxn UNIQUE (reactor_username, reaction_to);


--
-- Name: post_mentions no_dup_post_ment; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_mentions
    ADD CONSTRAINT no_dup_post_ment UNIQUE (post_id, username);


--
-- Name: post_reactions no_dup_post_rxn; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_reactions
    ADD CONSTRAINT no_dup_post_rxn UNIQUE (username, post_id);


--
-- Name: posts no_dup_repost; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT no_dup_repost UNIQUE (reposter_username, reposted_post_id);


--
-- Name: post_saves no_dup_saves; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_saves
    ADD CONSTRAINT no_dup_saves UNIQUE (username, post_id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id_);


--
-- Name: posts posts_pkey; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_pkey PRIMARY KEY (id_);


--
-- Name: chats ucu_pkey; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chats
    ADD CONSTRAINT ucu_pkey PRIMARY KEY (owner_user, partner_user);


--
-- Name: notifications unique_notif_key; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT unique_notif_key UNIQUE (notif_key);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (username);


--
-- Name: chat_history_entry_in_chat chat_history_entry_in_chat_che_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chat_history_entry_in_chat
    ADD CONSTRAINT chat_history_entry_in_chat_che_id_fkey FOREIGN KEY (che_id) REFERENCES public.chat_history_entry(id_) ON DELETE CASCADE;


--
-- Name: chat_history_entry chat_history_entry_reaction_to_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chat_history_entry
    ADD CONSTRAINT chat_history_entry_reaction_to_fkey FOREIGN KEY (reaction_to) REFERENCES public.chat_history_entry(id_) ON DELETE CASCADE;


--
-- Name: chat_history_entry chat_history_entry_reactor_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chat_history_entry
    ADD CONSTRAINT chat_history_entry_reactor_username_fkey FOREIGN KEY (reactor_username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: chat_history_entry chat_history_entry_reply_to_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chat_history_entry
    ADD CONSTRAINT chat_history_entry_reply_to_fkey FOREIGN KEY (reply_to) REFERENCES public.chat_history_entry(id_) ON DELETE CASCADE;


--
-- Name: chat_history_entry chat_history_entry_sender_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chat_history_entry
    ADD CONSTRAINT chat_history_entry_sender_username_fkey FOREIGN KEY (sender_username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: chats chats_owner_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chats
    ADD CONSTRAINT chats_owner_user_fkey FOREIGN KEY (owner_user) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: chats chats_partner_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chats
    ADD CONSTRAINT chats_partner_user_fkey FOREIGN KEY (partner_user) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: comment_mentions comment_mentions_comment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.comment_mentions
    ADD CONSTRAINT comment_mentions_comment_id_fkey FOREIGN KEY (comment_id) REFERENCES public.comments(comment_id) ON DELETE CASCADE;


--
-- Name: comment_mentions comment_mentions_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.comment_mentions
    ADD CONSTRAINT comment_mentions_username_fkey FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: comment_reactions comment_reactions_comment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.comment_reactions
    ADD CONSTRAINT comment_reactions_comment_id_fkey FOREIGN KEY (comment_id) REFERENCES public.comments(comment_id) ON DELETE CASCADE;


--
-- Name: comment_reactions comment_reactions_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.comment_reactions
    ADD CONSTRAINT comment_reactions_username_fkey FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: comments comments_parent_comment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_parent_comment_id_fkey FOREIGN KEY (parent_comment_id) REFERENCES public.comments(comment_id) ON DELETE CASCADE;


--
-- Name: comments comments_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id_) ON DELETE CASCADE;


--
-- Name: comments comments_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_username_fkey FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: follows follows_follower_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_follower_username_fkey FOREIGN KEY (follower_username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: follows follows_following_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_following_username_fkey FOREIGN KEY (following_username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: chat_history_entry_in_chat hist_in_chat; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.chat_history_entry_in_chat
    ADD CONSTRAINT hist_in_chat FOREIGN KEY (owner_user, partner_user) REFERENCES public.chats(owner_user, partner_user) ON DELETE CASCADE;


--
-- Name: notifications notifications_owner_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_owner_user_fkey FOREIGN KEY (owner_user) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: post_hashtags post_hashtags_htname_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_hashtags
    ADD CONSTRAINT post_hashtags_htname_fkey FOREIGN KEY (htname) REFERENCES public.hashtags(htname) ON DELETE CASCADE;


--
-- Name: post_hashtags post_hashtags_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_hashtags
    ADD CONSTRAINT post_hashtags_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id_) ON DELETE CASCADE;


--
-- Name: post_mentions post_mentions_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_mentions
    ADD CONSTRAINT post_mentions_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id_) ON DELETE CASCADE;


--
-- Name: post_mentions post_mentions_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_mentions
    ADD CONSTRAINT post_mentions_username_fkey FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: post_reactions post_reactions_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_reactions
    ADD CONSTRAINT post_reactions_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id_) ON DELETE CASCADE;


--
-- Name: post_reactions post_reactions_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_reactions
    ADD CONSTRAINT post_reactions_username_fkey FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: post_saves post_saves_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_saves
    ADD CONSTRAINT post_saves_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id_) ON DELETE CASCADE;


--
-- Name: post_saves post_saves_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.post_saves
    ADD CONSTRAINT post_saves_username_fkey FOREIGN KEY (username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: posts posts_owner_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_owner_user_fkey FOREIGN KEY (owner_user) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: posts posts_reposted_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_reposted_post_id_fkey FOREIGN KEY (reposted_post_id) REFERENCES public.posts(id_) ON DELETE CASCADE;


--
-- Name: posts posts_reposter_username_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9ine
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_reposter_username_fkey FOREIGN KEY (reposter_username) REFERENCES public.users(username) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict TLrEy1DfvQIg9iGPiLyMfsdNlBkvqeW0MIMKETZEDj9eUonkPIYb4YGPG4ee8aT

