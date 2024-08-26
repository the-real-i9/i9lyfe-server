--
-- PostgreSQL database dump
--

-- Dumped from database version 16.3 (Ubuntu 16.3-1.pgdg22.04+1)
-- Dumped by pg_dump version 16.3 (Ubuntu 16.3-1.pgdg22.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: i9l_user_profile_t; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.i9l_user_profile_t AS (
	user_id integer,
	username character varying,
	name character varying,
	bio character varying,
	profile_pic_url character varying,
	followers_count integer,
	following_count integer,
	client_follows boolean
);


ALTER TYPE public.i9l_user_profile_t OWNER TO postgres;

--
-- Name: i9l_user_t; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.i9l_user_t AS (
	id integer,
	email character varying,
	username character varying,
	name character varying,
	profile_pic_url character varying,
	connection_status text
);


ALTER TYPE public.i9l_user_t OWNER TO postgres;

--
-- Name: ui_comment_struct; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.ui_comment_struct AS (
	owner_user json,
	comment_id integer,
	comment_text text,
	attachment_url text,
	reactions_count integer,
	comments_count integer,
	client_reaction integer
);


ALTER TYPE public.ui_comment_struct OWNER TO postgres;

--
-- Name: ui_post_struct; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.ui_post_struct AS (
	owner_user json,
	post_id integer,
	type text,
	media_urls text[],
	description text,
	reactions_count integer,
	comments_count integer,
	reposts_count integer,
	saves_count integer,
	client_reaction integer,
	client_reposted boolean,
	client_saved boolean
);


ALTER TYPE public.ui_post_struct OWNER TO postgres;

--
-- Name: ack_msg_delivered(integer, integer, integer, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ack_msg_delivered(client_user_id integer, in_conversation_id integer, in_message_id integer, delivery_time timestamp without time zone) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
  convo_partner_user_id int;
BEGIN
IF (SELECT delivery_status FROM message_ WHERE id = in_message_id) <> 'delivered' THEN
  UPDATE user_conversation SET unread_messages_count = unread_messages_count + 1, updated_at = delivery_time
  WHERE user_id = client_user_id AND conversation_id = in_conversation_id
  RETURNING partner_user_id INTO convo_partner_user_id;
  
  -- convo_partner_user_id is a "guard" condition asserting that the message you ack must indeed belong to your conversation partner
  UPDATE message_ SET delivery_status = 'delivered'
  WHERE id = in_message_id AND conversation_id = in_conversation_id AND sender_user_id = convo_partner_user_id;
END IF;
  
  RETURN true;
END;
$$;


ALTER FUNCTION public.ack_msg_delivered(client_user_id integer, in_conversation_id integer, in_message_id integer, delivery_time timestamp without time zone) OWNER TO postgres;

--
-- Name: ack_msg_read(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ack_msg_read(client_user_id integer, in_conversation_id integer, in_message_id integer) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
  convo_partner_user_id int;
BEGIN
IF (SELECT delivery_status FROM message_ WHERE id = in_message_id) <> 'read' THEN
  UPDATE user_conversation SET unread_messages_count = CASE WHEN unread_messages_count > 0 THEN unread_messages_count - 1 ELSE 0 END
  WHERE user_id = client_user_id AND conversation_id = in_conversation_id
  RETURNING partner_user_id INTO convo_partner_user_id;
  
  -- convo_partner_user_id is a "guard" condition asserting that the message you ack must indeed belong to your conversation partner
  UPDATE message_ SET delivery_status = 'read'
  WHERE id = in_message_id AND conversation_id = in_conversation_id AND sender_user_id = convo_partner_user_id;
END IF;
 
  RETURN true;
END;
$$;


ALTER FUNCTION public.ack_msg_read(client_user_id integer, in_conversation_id integer, in_message_id integer) OWNER TO postgres;

--
-- Name: change_password(character varying, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.change_password(in_email character varying, in_new_password character varying) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  UPDATE i9l_user SET password = in_new_password WHERE email = in_email;
  
  RETURN true;
END;
$$;


ALTER FUNCTION public.change_password(in_email character varying, in_new_password character varying) OWNER TO postgres;

--
-- Name: create_comment_on_comment(integer, integer, integer, text, text, character varying[], character varying[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.create_comment_on_comment(OUT new_comment_id integer, OUT comment_notif json, OUT mention_notifs json[], OUT latest_comments_count integer, in_target_comment_id integer, target_comment_owner_user_id integer, client_user_id integer, in_comment_text text, in_attachment_url text, mentions character varying[], hashtags character varying[]) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  ret_comment_id int;
  
  mention_username varchar;
  ment_user_id int;
  
  client_data json;
  
  mention_notifs_acc json[] := ARRAY[]::json[];
  
  hashtag_n varchar;
BEGIN
  INSERT INTO comment_ (target_comment_id, commenter_user_id, comment_text, attachment_url)
  VALUES (in_target_comment_id, client_user_id, in_comment_text, in_attachment_url)
  RETURNING id INTO ret_comment_id;
  
  -- populate client data
  SELECT json_build_object(
	  'id', id,
	  'username', username,
	  'profile_pic_url', profile_pic_url
  ) INTO client_data FROM i9l_user WHERE id = client_user_id;
  
  
  FOREACH mention_username IN ARRAY mentions
  LOOP
	SELECT id INTO ment_user_id FROM i9l_user WHERE username = mention_username;

    -- skip if user doesn't exist
	CONTINUE WHEN ment_user_id is null;

	-- create mentions
    INSERT INTO pc_mention (comment_id, user_id)
	VALUES (ret_comment_id, ment_user_id);
	
	-- skip mention notification for client user
	CONTINUE WHEN ment_user_id = client_user_id;
	
	-- create mention notifications
	INSERT INTO notification (type, sender_user_id, receiver_user_id, via_comment_id)
	VALUES ('mention_in_comment', client_user_id, ment_user_id, ret_comment_id);
	
	mention_notifs_acc := array_append(mention_notifs_acc, json_build_object(
		'receiver_user_id', ment_user_id,
		'sender', client_data,
		'type', 'mention_in_comment',
		'comment_id', ret_comment_id
	));
  END LOOP;
  
  -- create hashtags
  FOREACH hashtag_n IN ARRAY hashtags
  LOOP
    INSERT INTO pc_hashtag (comment_id, hashtag_name)
	VALUES (ret_comment_id, hashtag_n);
  END LOOP;
  
  -- create comment notification
  INSERT INTO notification (type, sender_user_id, receiver_user_id, via_comment_id, comment_created_id)
  VALUES ('comment_on_comment', client_user_id, target_comment_owner_user_id, in_target_comment_id, ret_comment_id);
  
  
  new_comment_id := ret_comment_id;
  mention_notifs := mention_notifs_acc;
  comment_notif := json_build_object(
	  'receiver_user_id', target_comment_owner_user_id,
	  'type', 'comment_on_comment',
	  'sender', client_data,
	  'target_comment_id', in_target_comment_id,
	  'comment_created_id', ret_comment_id
  );
  
  SELECT COUNT(1) + 1 INTO latest_comments_count FROM comment_ WHERE target_comment_id = in_target_comment_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_comment_on_comment(OUT new_comment_id integer, OUT comment_notif json, OUT mention_notifs json[], OUT latest_comments_count integer, in_target_comment_id integer, target_comment_owner_user_id integer, client_user_id integer, in_comment_text text, in_attachment_url text, mentions character varying[], hashtags character varying[]) OWNER TO postgres;

--
-- Name: create_comment_on_post(integer, integer, integer, text, text, character varying[], character varying[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.create_comment_on_post(OUT new_comment_id integer, OUT comment_notif json, OUT mention_notifs json[], OUT latest_comments_count integer, in_target_post_id integer, target_post_owner_user_id integer, client_user_id integer, in_comment_text text, in_attachment_url text, mentions character varying[], hashtags character varying[]) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  ret_comment_id int;
  
  mention_username varchar;
  ment_user_id int;
  
  client_data json;
  
  mention_notifs_acc json[] := ARRAY[]::json[];
  
  hashtag_n varchar;
BEGIN
  INSERT INTO comment_ (target_post_id, commenter_user_id, comment_text, attachment_url)
  VALUES (in_target_post_id, client_user_id, in_comment_text, in_attachment_url)
  RETURNING id INTO ret_comment_id;
  
  -- populate client data
  SELECT json_build_object(
	  'id', id,
	  'username', username,
	  'profile_pic_url', profile_pic_url
  ) INTO client_data FROM i9l_user WHERE id = client_user_id;
  
  
  FOREACH mention_username IN ARRAY mentions
  LOOP
	SELECT id INTO ment_user_id FROM i9l_user WHERE username = mention_username;

    -- skip if user doesn't exist
	CONTINUE WHEN ment_user_id is null;

	-- create mentions
    INSERT INTO pc_mention (comment_id, user_id)
	VALUES (ret_comment_id, ment_user_id);
	
	-- skip mention notification for client user
	CONTINUE WHEN ment_user_id = client_user_id;
	
	-- create mention notifications
	INSERT INTO notification (type, sender_user_id, receiver_user_id, via_comment_id)
	VALUES ('mention_in_comment', client_user_id, ment_user_id, ret_comment_id);
	
	mention_notifs_acc := array_append(mention_notifs_acc, json_build_object(
		'receiver_user_id', ment_user_id,
		'sender', client_data,
		'type', 'mention_in_comment',
		'comment_id', ret_comment_id
	));
  END LOOP;
  
  -- create hashtags
  FOREACH hashtag_n IN ARRAY hashtags
  LOOP
    INSERT INTO pc_hashtag (comment_id, hashtag_name)
	VALUES (ret_comment_id, hashtag_n);
  END LOOP;
  
  -- create comment notification
  INSERT INTO notification (type, sender_user_id, receiver_user_id, via_post_id, comment_created_id)
  VALUES ('comment_on_post', client_user_id, target_post_owner_user_id, in_target_post_id, ret_comment_id);
  
  
  new_comment_id := ret_comment_id;
  mention_notifs := mention_notifs_acc;
  comment_notif := json_build_object(
	  'receiver_user_id', target_post_owner_user_id,
	  'type', 'comment_on_post',
	  'sender', client_data,
	  'post_id', in_target_post_id,
	  'comment_created_id', ret_comment_id
  );
  
  SELECT COUNT(1) + 1 INTO latest_comments_count FROM comment_ WHERE target_post_id = in_target_post_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_comment_on_post(OUT new_comment_id integer, OUT comment_notif json, OUT mention_notifs json[], OUT latest_comments_count integer, in_target_post_id integer, target_post_owner_user_id integer, client_user_id integer, in_comment_text text, in_attachment_url text, mentions character varying[], hashtags character varying[]) OWNER TO postgres;

--
-- Name: create_conversation(integer, integer, json); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.create_conversation(OUT client_res json, OUT partner_res json, in_initiator_user_id integer, in_with_user_id integer, init_message json) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  ret_conversation_id int;
  ret_message_id int;
  
  client_data json;
BEGIN
  INSERT INTO conversation(initiator_user_id, with_user_id)
  VALUES (in_initiator_user_id, in_with_user_id)
  RETURNING id INTO ret_conversation_id;
  
  INSERT INTO user_conversation(conversation_id, user_id, partner_user_id)
  VALUES (ret_conversation_id, in_initiator_user_id, in_with_user_id);
  
  INSERT INTO user_conversation(conversation_id, user_id, partner_user_id)
  VALUES (ret_conversation_id, in_with_user_id, in_initiator_user_id);
  
  INSERT INTO message_(sender_user_id, conversation_id, msg_content)
  VALUES (in_initiator_user_id, ret_conversation_id, init_message)
  RETURNING id INTO ret_message_id;
  
  SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) INTO client_data
  FROM i9l_user WHERE id = in_initiator_user_id;
  
  client_res := json_build_object('conversation_id', ret_conversation_id, 'init_message_id', ret_message_id);
  
  partner_res := json_build_object(
	  'conversation', json_build_object(
		  'id', ret_conversation_id,
		  'partner', client_data
	  ),
	  'init_message', json_build_object(
		  'id', ret_message_id,
		  'sender', client_data,
		  'msg_content', init_message
	  )
  );
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_conversation(OUT client_res json, OUT partner_res json, in_initiator_user_id integer, in_with_user_id integer, init_message json) OWNER TO postgres;

--
-- Name: create_message(integer, integer, json); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.create_message(OUT client_res json, OUT partner_res json, in_conversation_id integer, client_user_id integer, in_msg_content json) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  ret_msg_id int;
  
  client_data json;
BEGIN
  INSERT INTO message_ (sender_user_id, conversation_id, msg_content) 
  VALUES (client_user_id, in_conversation_id, in_msg_content)
  RETURNING id INTO ret_msg_id;
  
  -- sender data
  SELECT json_build_object(
	  'id', id,
	  'username', username,
	  'profile_pic_url', profile_pic_url,
	  'connection_status', connection_status
  ) INTO client_data
  FROM i9l_user WHERE id = client_user_id;
  
  client_res := json_build_object('new_msg_id', ret_msg_id);
  
  partner_res := json_build_object(
	  'conversation_id', in_conversation_id,
	  'new_msg_id', ret_msg_id,
	  'sender', client_data,
	  'msg_content', in_msg_content
  );
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_message(OUT client_res json, OUT partner_res json, in_conversation_id integer, client_user_id integer, in_msg_content json) OWNER TO postgres;

--
-- Name: create_post(integer, text[], text, text, character varying[], character varying[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.create_post(OUT new_post_id integer, OUT mention_notifs json[], client_user_id integer, in_media_urls text[], in_type text, in_description text, mentions character varying[], hashtags character varying[]) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  ret_post_id int;
  
  mention_username varchar;
  ment_user_id int;
  
  client_data json;
  
  mention_notifs_acc json[] := ARRAY[]::json[];
  
  hashtag_n varchar;
BEGIN
  INSERT INTO post (user_id, type, media_urls, description)
  VALUES (client_user_id, in_type, in_media_urls, in_description)
  RETURNING id INTO ret_post_id;
  
  -- populate client data
  SELECT json_build_object(
	  'id', id,
	  'username', username,
	  'profile_pic_url', profile_pic_url
  ) INTO client_data FROM i9l_user WHERE id = client_user_id;
  
  
  FOREACH mention_username IN ARRAY mentions
  LOOP
	SELECT id INTO ment_user_id FROM i9l_user WHERE username = mention_username;

    -- skip if mentioned user is not found
	CONTINUE WHEN ment_user_id is null;

	-- create mentions
    INSERT INTO pc_mention (post_id, user_id)
	VALUES (ret_post_id, ment_user_id);
	
	-- skip mention notification for client user
	CONTINUE WHEN ment_user_id = client_user_id;
	
	-- create mention notifications
	INSERT INTO notification (type, sender_user_id, receiver_user_id, via_post_id)
	VALUES ('mention_in_post', client_user_id, ment_user_id, ret_post_id);
	
	mention_notifs_acc := array_append(mention_notifs_acc, json_build_object(
		'receiver_user_id', ment_user_id,
		'sender', client_data,
		'type', 'mention_in_post',
		'post_id', ret_post_id
	));
  END LOOP;
  
  
  -- create hashtags
  
  FOREACH hashtag_n IN ARRAY hashtags
  LOOP
    INSERT INTO pc_hashtag (post_id, hashtag_name)
	VALUES (ret_post_id, hashtag_n);
  END LOOP;
  
  
  
  new_post_id := ret_post_id;
  mention_notifs := mention_notifs_acc;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_post(OUT new_post_id integer, OUT mention_notifs json[], client_user_id integer, in_media_urls text[], in_type text, in_description text, mentions character varying[], hashtags character varying[]) OWNER TO postgres;

--
-- Name: create_reaction_to_comment(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.create_reaction_to_comment(OUT reaction_notif json, OUT latest_reactions_count integer, client_user_id integer, in_target_comment_id integer, target_comment_owner_user_id integer, in_reaction_code_point integer) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  client_data json;
BEGIN
  INSERT INTO pc_reaction (reactor_user_id, target_comment_id, reaction_code_point)
  VALUES (client_user_id, in_target_comment_id, in_reaction_code_point);
  
  -- populate client data
  SELECT json_build_object(
	  'id', id,
	  'username', username,
	  'profile_pic_url', profile_pic_url
  ) INTO client_data FROM i9l_user WHERE id = client_user_id;
  
  INSERT INTO notification (type, sender_user_id, receiver_user_id, via_comment_id, reaction_code_point)
  VALUES ('reaction_to_comment', client_user_id, target_comment_owner_user_id, in_target_comment_id, in_reaction_code_point);
  
  reaction_notif := json_build_object(
	  'receiver_user_id', target_comment_owner_user_id,
	  'type', 'reaction_to_comment',
	  'reaction_code_point', in_reaction_code_point,
	  'comment_id', in_target_comment_id,
	  'sender', client_data
	  
  );
  
  SELECT COUNT(1) + 1 INTO latest_reactions_count FROM pc_reaction WHERE target_comment_id = in_target_comment_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_reaction_to_comment(OUT reaction_notif json, OUT latest_reactions_count integer, client_user_id integer, in_target_comment_id integer, target_comment_owner_user_id integer, in_reaction_code_point integer) OWNER TO postgres;

--
-- Name: create_reaction_to_post(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.create_reaction_to_post(OUT reaction_notif json, OUT latest_reactions_count integer, client_user_id integer, in_target_post_id integer, target_post_owner_user_id integer, in_reaction_code_point integer) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  client_data json;
BEGIN
  INSERT INTO pc_reaction (reactor_user_id, target_post_id, reaction_code_point)
  VALUES (client_user_id, in_target_post_id, in_reaction_code_point);
  
  -- populate client data
  SELECT json_build_object(
	  'id', id,
	  'username', username,
	  'profile_pic_url', profile_pic_url
  ) INTO client_data FROM i9l_user WHERE id = client_user_id;
  
  INSERT INTO notification (type, sender_user_id, receiver_user_id, via_post_id, reaction_code_point)
  VALUES ('reaction_to_post', client_user_id, target_post_owner_user_id, in_target_post_id, in_reaction_code_point);
  
  reaction_notif := json_build_object(
	  'receiver_user_id', target_post_owner_user_id,
	  'type', 'reaction_to_post',
	  'post_id', in_target_post_id,
	  'reaction_code_point', in_reaction_code_point,
	  'sender', client_data
  );
  
  SELECT COUNT(1) + 1 INTO latest_reactions_count FROM pc_reaction WHERE target_post_id = in_target_post_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_reaction_to_post(OUT reaction_notif json, OUT latest_reactions_count integer, client_user_id integer, in_target_post_id integer, target_post_owner_user_id integer, in_reaction_code_point integer) OWNER TO postgres;

--
-- Name: create_user(character varying, character varying, character varying, character varying, timestamp without time zone, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.create_user(in_email character varying, in_username character varying, in_password character varying, in_name character varying, in_birthday timestamp without time zone, in_bio character varying) RETURNS SETOF public.i9l_user_t
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY 
	INSERT INTO i9l_user(email, username, password, name, birthday, bio)
    VALUES(in_email, in_username, in_password, in_name, in_birthday, in_bio) 
    RETURNING id, email, username, name, profile_pic_url, connection_status;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_user(in_email character varying, in_username character varying, in_password character varying, in_name character varying, in_birthday timestamp without time zone, in_bio character varying) OWNER TO postgres;

--
-- Name: edit_user(integer, character varying[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.edit_user(client_user_id integer, col_updates character varying[]) RETURNS boolean
    LANGUAGE plpgsql
    AS $_$
DECLARE 
  col_name_val varchar[];
  update_sets varchar := '';
BEGIN
  FOREACH col_name_val SLICE 1 IN ARRAY col_updates
  LOOP
    IF col_name_val[1] NOT IN ('name', 'birthday', 'bio') THEN
	  RAISE EXCEPTION '"%" is either an invalid or a non-editable column', col_name_val[1] 
	  USING HINT = 'Validate column name or set column from the proper routine';
	END IF;
    update_sets := update_sets || col_name_val[1] || ' = ''' || col_name_val[2] || ''', ';
  END LOOP;
  
  update_sets := LEFT(update_sets, LENGTH(update_sets) - 2);
  
  EXECUTE 'UPDATE i9l_user SET ' || update_sets || ' WHERE id = $1' USING client_user_id;
  
  RETURN true;
END;
$_$;


ALTER FUNCTION public.edit_user(client_user_id integer, col_updates character varying[]) OWNER TO postgres;

--
-- Name: follow_user(integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.follow_user(OUT follow_notif json, client_user_id integer, to_follow_user_id integer) RETURNS json
    LANGUAGE plpgsql
    AS $$
DECLARE
  client_data json;
BEGIN
  -- create follow relationship
  INSERT INTO follow (follower_user_id, followee_user_id) 
  VALUES (client_user_id, to_follow_user_id);
	  
  -- create follow notification
  INSERT INTO notification (type, sender_user_id, receiver_user_id) 
  VALUES ('follow', client_user_id, to_follow_user_id);
  
  -- populate client_data
  SELECT json_build_object(
	  'user_id', id,
	  'username', username,
	  'profile_pic_url', profile_pic_url
  ) INTO client_data FROM i9l_user WHERE id = client_user_id;
	  
  -- create and assign notification_data
  follow_notif := json_build_object(
	  'type', 'follow',
	  'receiver_user_id', to_follow_user_id,
	  'sender', client_data
  );
  
  RETURN;
END;
$$;


ALTER FUNCTION public.follow_user(OUT follow_notif json, client_user_id integer, to_follow_user_id integer) OWNER TO postgres;

--
-- Name: get_comment(integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_comment(in_comment_id integer, client_user_id integer) RETURNS SETOF public.ui_comment_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      comment_id,
      comment_text,
      attachment_url,
      reactions_count,
      comments_count,
      CASE 
        WHEN reactor_user_id = client_user_id THEN reaction_code_point 
        ELSE NULL
      END AS client_reaction
    FROM "CommentView"
    WHERE comment_id = in_comment_id;
	  
END;
$$;


ALTER FUNCTION public.get_comment(in_comment_id integer, client_user_id integer) OWNER TO postgres;

--
-- Name: get_comments_on_comment(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_comments_on_comment(in_target_comment_id integer, client_user_id integer, in_limit integer, in_offset integer) RETURNS SETOF public.ui_comment_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      comment_id,
      comment_text,
      attachment_url,
      reactions_count,
      comments_count,
      CASE 
        WHEN reactor_user_id = client_user_id THEN reaction_code_point 
        ELSE NULL
      END AS client_reaction
    FROM "CommentView"
    WHERE target_comment_id = in_target_comment_id
    ORDER BY created_at DESC
    LIMIT in_limit OFFSET in_offset;
	  
END;
$$;


ALTER FUNCTION public.get_comments_on_comment(in_target_comment_id integer, client_user_id integer, in_limit integer, in_offset integer) OWNER TO postgres;

--
-- Name: get_comments_on_post(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_comments_on_post(in_target_post_id integer, client_user_id integer, in_limit integer, in_offset integer) RETURNS SETOF public.ui_comment_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      comment_id,
      comment_text,
      attachment_url,
      reactions_count,
      comments_count,
      CASE 
        WHEN reactor_user_id = client_user_id THEN reaction_code_point 
        ELSE NULL
      END AS client_reaction
    FROM "CommentView"
    WHERE target_post_id = in_target_post_id
    ORDER BY created_at DESC
    LIMIT in_limit OFFSET in_offset;
	  
END;
$$;


ALTER FUNCTION public.get_comments_on_post(in_target_post_id integer, client_user_id integer, in_limit integer, in_offset integer) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: i9l_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.i9l_user (
    id integer NOT NULL,
    email character varying(255) NOT NULL,
    username character varying(255) NOT NULL,
    password character varying NOT NULL,
    name character varying NOT NULL,
    birthday date NOT NULL,
    bio character varying(300) DEFAULT 'Hey there! I"m using i9lyfe.'::character varying,
    profile_pic_url character varying DEFAULT ''::character varying NOT NULL,
    connection_status text DEFAULT 'online'::text NOT NULL,
    last_active timestamp without time zone,
    acc_deleted boolean DEFAULT false,
    cover_pic_url text DEFAULT ''::text NOT NULL,
    CONSTRAINT "User_check" CHECK ((((connection_status = 'offline'::text) AND (last_active IS NOT NULL)) OR ((connection_status = 'online'::text) AND (last_active IS NULL)))),
    CONSTRAINT "User_connection_status_check" CHECK ((connection_status = ANY (ARRAY['online'::text, 'offline'::text])))
);


ALTER TABLE public.i9l_user OWNER TO postgres;

--
-- Name: message_; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.message_ (
    id integer NOT NULL,
    sender_user_id integer NOT NULL,
    conversation_id integer NOT NULL,
    msg_content jsonb,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    delivery_status text DEFAULT 'sent'::text NOT NULL,
    reply_to_id integer,
    CONSTRAINT "Message_delivery_status_check" CHECK ((delivery_status = ANY (ARRAY['sent'::text, 'delivered'::text, 'read'::text])))
);


ALTER TABLE public.message_ OWNER TO postgres;

--
-- Name: message_reaction; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.message_reaction (
    id integer NOT NULL,
    message_id integer NOT NULL,
    reactor_user_id integer NOT NULL,
    reaction_code_point integer NOT NULL
);


ALTER TABLE public.message_reaction OWNER TO postgres;

--
-- Name: ConversationHistoryView; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public."ConversationHistoryView" AS
 SELECT msg.id AS msg_id,
    json_build_object('id', sender.id, 'username', sender.username, 'profile_pic_url', sender.profile_pic_url) AS sender,
    msg.msg_content,
    msg.delivery_status,
    ( SELECT array_agg(message_reaction.reaction_code_point) AS array_agg
           FROM public.message_reaction
          WHERE (message_reaction.message_id = msg.id)) AS reactions,
    msg.created_at,
    msg.conversation_id
   FROM (public.message_ msg
     JOIN public.i9l_user sender ON ((sender.id = msg.sender_user_id)))
  ORDER BY msg.created_at DESC;


ALTER VIEW public."ConversationHistoryView" OWNER TO postgres;

--
-- Name: get_conversation_history(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_conversation_history(in_conversation_id integer, in_limit integer, in_offset integer) RETURNS SETOF public."ConversationHistoryView"
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT * FROM (
	  SELECT * FROM "ConversationHistoryView"
      WHERE conversation_id = in_conversation_id
      LIMIT in_limit OFFSET in_offset
  ) ORDER BY created_at ASC;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_conversation_history(in_conversation_id integer, in_limit integer, in_offset integer) OWNER TO postgres;

--
-- Name: get_explore_posts(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_explore_posts(in_limit integer, in_offset integer, client_user_id integer) RETURNS SETOF public.ui_post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = client_user_id THEN reaction_code_point
        ELSE NULL
      END client_reaction,
      CASE 
        WHEN reposter_user_id = client_user_id THEN true
        ELSE false
      END client_reposted,
      CASE 
        WHEN saver_user_id = client_user_id THEN true
        ELSE false
      END client_saved
    FROM "PostView"
    ORDER BY created_at DESC
	LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.get_explore_posts(in_limit integer, in_offset integer, client_user_id integer) OWNER TO postgres;

--
-- Name: get_feed_posts(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_feed_posts(client_user_id integer, in_limit integer, in_offset integer) RETURNS SETOF public.ui_post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = client_user_id THEN reaction_code_point
        ELSE NULL
      END client_reaction,
      CASE 
        WHEN reposter_user_id = client_user_id THEN true
        ELSE false
      END client_reposted,
      CASE 
        WHEN saver_user_id = client_user_id THEN true
        ELSE false
      END client_saved
    FROM "PostView"
    LEFT JOIN follow ON follow.followee_user_id = owner_user_id
    WHERE follow.follower_user_id = client_user_id OR owner_user_id = client_user_id
    ORDER BY created_at DESC
    LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.get_feed_posts(client_user_id integer, in_limit integer, in_offset integer) OWNER TO postgres;

--
-- Name: get_hashtag_posts(character varying, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_hashtag_posts(in_hashtag_name character varying, in_limit integer, in_offset integer, client_user_id integer) RETURNS SETOF public.ui_post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      pv.post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = client_user_id THEN reaction_code_point
        ELSE NULL
      END client_reaction,
      CASE 
        WHEN reposter_user_id = client_user_id THEN true
        ELSE false
      END client_reposted,
      CASE 
        WHEN saver_user_id = client_user_id THEN true
        ELSE false
      END client_saved
    FROM "PostView" pv
    INNER JOIN pc_hashtag pch ON pch.post_id = pv.post_id AND pch.hashtag_name = in_hashtag_name
	ORDER BY pv.created_at DESC
	LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.get_hashtag_posts(in_hashtag_name character varying, in_limit integer, in_offset integer, client_user_id integer) OWNER TO postgres;

--
-- Name: get_mentioned_posts(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_mentioned_posts(in_limit integer, in_offset integer, client_user_id integer) RETURNS SETOF public.ui_post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      pv.post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = client_user_id THEN reaction_code_point
        ELSE NULL
      END client_reaction,
      CASE 
        WHEN reposter_user_id = client_user_id THEN true
        ELSE false
      END client_reposted,
      CASE 
        WHEN saver_user_id = client_user_id THEN true
        ELSE false
      END client_saved
    FROM "PostView" pv
    INNER JOIN pc_mention ON pc_mention.post_id = pv.post_id AND pc_mention.user_id = client_user_id
	ORDER BY pv.created_at DESC
	LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.get_mentioned_posts(in_limit integer, in_offset integer, client_user_id integer) OWNER TO postgres;

--
-- Name: get_post(integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_post(in_post_id integer, client_user_id integer) RETURNS SETOF public.ui_post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = client_user_id THEN reaction_code_point
        ELSE NULL
      END client_reaction,
      CASE 
        WHEN reposter_user_id = client_user_id THEN true
        ELSE false
      END client_reposted,
      CASE 
        WHEN saver_user_id = client_user_id THEN true
        ELSE false
      END client_saved
    FROM "PostView"
    WHERE post_id = in_post_id;
	  
	  
END;
$$;


ALTER FUNCTION public.get_post(in_post_id integer, client_user_id integer) OWNER TO postgres;

--
-- Name: get_reacted_posts(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_reacted_posts(in_limit integer, in_offset integer, client_user_id integer) RETURNS SETOF public.ui_post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = client_user_id THEN reaction_code_point
        ELSE NULL
      END client_reaction,
      CASE 
        WHEN reposter_user_id = client_user_id THEN true
        ELSE false
      END client_reposted,
      CASE 
        WHEN saver_user_id = client_user_id THEN true
        ELSE false
      END client_saved
    FROM "PostView"
    WHERE reactor_user_id = client_user_id
	ORDER BY created_at DESC
	LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.get_reacted_posts(in_limit integer, in_offset integer, client_user_id integer) OWNER TO postgres;

--
-- Name: get_reactors_to_comment(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_reactors_to_comment(in_comment_id integer, client_user_id integer, in_limit integer, in_offset integer) RETURNS TABLE(reactor_user json, client_follows boolean)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
	    'id', i9l_user.id, 
        'profile_pic_url', i9l_user.profile_pic_url, 
        'username', i9l_user.username, 
        'name', i9l_user.name
	  ) AS reactor_user,
      CASE
        WHEN client_follows.id IS NULL THEN false
        ELSE true
      END client_follows
    FROM pc_reaction 
    INNER JOIN i9l_user i9l_user ON pc_reaction.reactor_user_id = i9l_user.id 
    LEFT JOIN follow client_follows ON client_follows.followee_user_id = i9l_user.id AND client_follows.follower_user_id = client_user_id
    WHERE pc_reaction.target_comment_id = in_comment_id
    ORDER BY pc_reaction.created_at DESC
    LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.get_reactors_to_comment(in_comment_id integer, client_user_id integer, in_limit integer, in_offset integer) OWNER TO postgres;

--
-- Name: get_reactors_to_post(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_reactors_to_post(in_post_id integer, client_user_id integer, in_limit integer, in_offset integer) RETURNS TABLE(reactor_user json, client_follows boolean)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
	    'id', i9l_user.id, 
        'profile_pic_url', i9l_user.profile_pic_url, 
        'username', i9l_user.username, 
        'name', i9l_user.name
	  ) AS reactor_user,
      CASE
        WHEN client_follows.id IS NULL THEN false
        ELSE true
      END client_follows
    FROM pc_reaction 
    INNER JOIN i9l_user i9l_user ON pc_reaction.reactor_user_id = i9l_user.id 
    LEFT JOIN follow client_follows ON client_follows.followee_user_id = i9l_user.id AND client_follows.follower_user_id = client_user_id
    WHERE pc_reaction.target_post_id = in_post_id
    ORDER BY pc_reaction.created_at DESC
    LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.get_reactors_to_post(in_post_id integer, client_user_id integer, in_limit integer, in_offset integer) OWNER TO postgres;

--
-- Name: get_reactors_with_reaction_to_comment(integer, integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_reactors_with_reaction_to_comment(in_comment_id integer, in_reaction_code_point integer, client_user_id integer, in_limit integer, in_offset integer) RETURNS TABLE(reactor_user json, client_follows boolean)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
	    'id', i9l_user.id, 
        'profile_pic_url', i9l_user.profile_pic_url, 
        'username', i9l_user.username, 
        'name', i9l_user.name
	  ) AS reactor_user,
      CASE
        WHEN client_follows.id IS NULL THEN false
        ELSE true
      END client_follows
    FROM pc_reaction 
    INNER JOIN i9l_user i9l_user ON pc_reaction.reactor_user_id = i9l_user.id 
    LEFT JOIN follow client_follows ON client_follows.followee_user_id = i9l_user.id AND client_follows.follower_user_id = client_user_id
    WHERE pc_reaction.target_comment_id = in_comment_id AND pc_reaction.reaction_code_point = in_reaction_code_point
    ORDER BY pc_reaction.created_at DESC
    LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.get_reactors_with_reaction_to_comment(in_comment_id integer, in_reaction_code_point integer, client_user_id integer, in_limit integer, in_offset integer) OWNER TO postgres;

--
-- Name: get_reactors_with_reaction_to_post(integer, integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_reactors_with_reaction_to_post(in_post_id integer, in_reaction_code_point integer, client_user_id integer, in_limit integer, in_offset integer) RETURNS TABLE(reactor_user json, client_follows boolean)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
	    'id', i9l_user.id, 
        'profile_pic_url', i9l_user.profile_pic_url, 
        'username', i9l_user.username, 
        'name', i9l_user.name
	  ) AS reactor_user,
      CASE
        WHEN client_follows.id IS NULL THEN false
        ELSE true
      END client_follows
    FROM pc_reaction 
    INNER JOIN i9l_user i9l_user ON pc_reaction.reactor_user_id = i9l_user.id 
    LEFT JOIN follow client_follows ON client_follows.followee_user_id = i9l_user.id AND client_follows.follower_user_id = client_user_id
    WHERE pc_reaction.target_post_id = in_post_id AND pc_reaction.reaction_code_point = in_reaction_code_point
    ORDER BY pc_reaction.created_at DESC
    LIMIT in_limit OFFSET in_offset;
	
	END;
$$;


ALTER FUNCTION public.get_reactors_with_reaction_to_post(in_post_id integer, in_reaction_code_point integer, client_user_id integer, in_limit integer, in_offset integer) OWNER TO postgres;

--
-- Name: get_saved_posts(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_saved_posts(in_limit integer, in_offset integer, client_user_id integer) RETURNS SETOF public.ui_post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = client_user_id THEN reaction_code_point
        ELSE NULL
      END client_reaction,
      CASE 
        WHEN reposter_user_id = client_user_id THEN true
        ELSE false
      END client_reposted,
      CASE 
        WHEN saver_user_id = client_user_id THEN true
        ELSE false
      END client_saved
    FROM "PostView"
    WHERE saver_user_id = client_user_id
	ORDER BY created_at DESC
	LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.get_saved_posts(in_limit integer, in_offset integer, client_user_id integer) OWNER TO postgres;

--
-- Name: get_user(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_user(unique_identifier character varying) RETURNS SETOF public.i9l_user_t
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
	SELECT id, email, username, name, profile_pic_url, connection_status
	FROM i9l_user
    WHERE unique_identifier = ANY(ARRAY[id::varchar, email, username]);
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_user(unique_identifier character varying) OWNER TO postgres;

--
-- Name: get_user_conversations(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_user_conversations(client_user_id integer) RETURNS TABLE(conversation_id integer, partner json, unread_messages_count integer, updated_at timestamp without time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT uconv.conversation_id,
    json_build_object(
		'id', par.id,
		'username', par.username,
		'profile_pic_url', par.profile_pic_url,
		'connection_status', par.connection_status,
		'last_active', par.last_active
	) AS partner,
    uconv.unread_messages_count,
    uconv.updated_at
  FROM user_conversation uconv
  LEFT JOIN i9l_user par ON par.id = uconv.partner_user_id
  WHERE uconv.user_id = client_user_id AND uconv.deleted = false;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_user_conversations(client_user_id integer) OWNER TO postgres;

--
-- Name: get_user_followers(character varying, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_user_followers(in_username character varying, in_limit integer, in_offset integer, client_user_id integer) RETURNS TABLE(user_id integer, username character varying, bio character varying, profile_pic_url character varying, client_follows boolean)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT follower_user.id AS user_id, 
      follower_user.username, 
      follower_user.bio, 
      follower_user.profile_pic_url,
      CASE
        WHEN client_follows.id IS NULL THEN false
        ELSE true
      END client_follows
    FROM follow
    LEFT JOIN i9l_user follower_user ON follower_user.id = follow.follower_user_id
    LEFT JOIN i9l_user followee_user ON followee_user.id = follow.followee_user_id
    LEFT JOIN follow client_follows 
      ON client_follows.followee_user_id = follower_user.id AND client_follows.follower_user_id = client_user_id
    WHERE followee_user.username = in_username
	ORDER BY follow.follow_on DESC
	LIMIT in_limit OFFSET in_offset;
	
	
	END;
$$;


ALTER FUNCTION public.get_user_followers(in_username character varying, in_limit integer, in_offset integer, client_user_id integer) OWNER TO postgres;

--
-- Name: get_user_following(character varying, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_user_following(in_username character varying, in_limit integer, in_offset integer, client_user_id integer) RETURNS TABLE(user_id integer, username character varying, bio character varying, profile_pic_url character varying, client_follows boolean)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT followee_user.id AS user_id, 
      followee_user.username, 
      followee_user.bio, 
      followee_user.profile_pic_url,
      CASE
        WHEN client_follows.id IS NULL THEN false
        ELSE true
      END client_follows
    FROM follow
    LEFT JOIN i9l_user follower_user ON follower_user.id = follow.follower_user_id
    LEFT JOIN i9l_user followee_user ON followee_user.id = follow.followee_user_id
    LEFT JOIN follow client_follows 
      ON client_follows.followee_user_id = followee_user.id AND client_follows.follower_user_id = client_user_id
    WHERE follower_user.username = in_username
	ORDER BY follow.follow_on DESC
	LIMIT in_limit OFFSET in_offset;
	  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_user_following(in_username character varying, in_limit integer, in_offset integer, client_user_id integer) OWNER TO postgres;

--
-- Name: get_user_notifications(integer, timestamp without time zone, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_user_notifications(client_user_id integer, in_from timestamp without time zone, in_limit integer, in_offset integer) RETURNS SETOF json
    LANGUAGE plpgsql
    AS $$
BEGIN
	RETURN QUERY
    SELECT json_strip_nulls(json_build_object(
	  'type', n.type,
	  'is_read', n.is_read,
	  'sender', json_build_object(
		  'id', sender.id,
		  'username', sender.username,
		  'profile_pic_url', sender.profile_pic_url
	  ),
	  'via_post_id', n.via_post_id,
	  'via_comment_id', n.via_comment_id,
	  'comment_created_id', n.comment_created_id,
	  'reaction_code_point', n.reaction_code_point,
	  'created_at', n.created_at
  )) FROM notification n
  INNER JOIN i9l_user sender ON sender.id = n.sender_user_id
  WHERE n.receiver_user_id = client_user_id AND n.created_at >= in_from
  ORDER BY n.created_at DESC
  LIMIT in_limit OFFSET in_offset;
  
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_user_notifications(client_user_id integer, in_from timestamp without time zone, in_limit integer, in_offset integer) OWNER TO postgres;

--
-- Name: get_user_password(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_user_password(OUT pswd character varying, unique_identifier character varying) RETURNS character varying
    LANGUAGE plpgsql
    AS $$
BEGIN
    SELECT password FROM i9l_user
	WHERE unique_identifier = ANY(ARRAY[id::varchar, email, username])
	INTO pswd;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_user_password(OUT pswd character varying, unique_identifier character varying) OWNER TO postgres;

--
-- Name: get_user_posts(character varying, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_user_posts(in_username character varying, in_limit integer, in_offset integer, client_user_id integer) RETURNS SETOF public.ui_post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = client_user_id THEN reaction_code_point
        ELSE NULL
      END client_reaction,
      CASE 
        WHEN reposter_user_id = client_user_id THEN true
        ELSE false
      END client_reposted,
      CASE 
        WHEN saver_user_id = client_user_id THEN true
        ELSE false
      END client_saved
    FROM "PostView"
    WHERE owner_username = in_username
	ORDER BY created_at DESC
	LIMIT in_limit OFFSET in_offset;

END;
$$;


ALTER FUNCTION public.get_user_posts(in_username character varying, in_limit integer, in_offset integer, client_user_id integer) OWNER TO postgres;

--
-- Name: get_user_profile(character varying, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_user_profile(in_username character varying, client_user_id integer) RETURNS SETOF public.i9l_user_profile_t
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY 
	SELECT i9l_user.id, username, name, bio, profile_pic_url, 
      COUNT(followee.id)::int, 
      COUNT(follower.id)::int,
      CASE
        WHEN client_follows.id IS NULL THEN false
        ELSE true
      END
    FROM i9l_user
    LEFT JOIN follow followee ON followee.followee_user_id = i9l_user.id
    LEFT JOIN follow follower ON follower.follower_user_id = i9l_user.id
    LEFT JOIN follow client_follows 
      ON client_follows.followee_user_id = i9l_user.id AND client_follows.follower_user_id = client_user_id
    WHERE i9l_user.username = in_username
    GROUP BY i9l_user.id,
      name,
      username,
      bio,
      profile_pic_url,
      client_follows.id;
	  
	  
END;
$$;


ALTER FUNCTION public.get_user_profile(in_username character varying, client_user_id integer) OWNER TO postgres;

--
-- Name: get_users_to_chat(text, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_users_to_chat(in_search text, in_limit integer, in_offset integer, client_user_id integer) RETURNS TABLE(id integer, username character varying, name character varying, profile_pic_url character varying, connection_status text, conversation_id integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT i9l_user.id,
	  i9l_user.username,
	  i9l_user.name,
	  i9l_user.profile_pic_url,
	  i9l_user.connection_status,
	  uconv.conversation_id
  FROM i9l_user
  LEFT JOIN user_conversation uconv ON uconv.user_id = i9l_user.id
  WHERE (i9l_user.username ILIKE in_search OR i9l_user.name ILIKE in_search) AND i9l_user.id != client_user_id
  LIMIT in_limit OFFSET in_offset;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_users_to_chat(in_search text, in_limit integer, in_offset integer, client_user_id integer) OWNER TO postgres;

--
-- Name: search_filter_posts(text, text, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.search_filter_posts(search_text text, filter_text text, in_limit integer, in_offset integer, client_user_id integer) RETURNS SETOF public.ui_post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = client_user_id THEN reaction_code_point
        ELSE NULL
      END client_reaction,
      CASE 
        WHEN reposter_user_id = client_user_id THEN true
        ELSE false
      END client_reposted,
      CASE 
        WHEN saver_user_id = client_user_id THEN true
        ELSE false
      END client_saved
    FROM "PostView"
    WHERE CASE 
	  WHEN filter_text <> 'all' THEN (to_tsvector(description) @@ to_tsquery(search_text) AND type = filter_text) 
	  ELSE to_tsvector(description) @@ to_tsquery(search_text) 
	END
	ORDER BY created_at DESC
	LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.search_filter_posts(search_text text, filter_text text, in_limit integer, in_offset integer, client_user_id integer) OWNER TO postgres;

--
-- Name: user_exists(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.user_exists(OUT check_res boolean, unique_identifier character varying) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT EXISTS(SELECT 1 
				FROM i9l_user 
				WHERE unique_identifier = ANY(ARRAY[id::varchar, email, username])
				)
  INTO check_res;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.user_exists(OUT check_res boolean, unique_identifier character varying) OWNER TO postgres;

--
-- Name: comment_; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.comment_ (
    id integer NOT NULL,
    comment_text text NOT NULL,
    commenter_user_id integer NOT NULL,
    attachment_url text,
    target_post_id integer,
    target_comment_id integer,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT either_comment_on_post_or_reply_to_comment CHECK (((target_post_id IS NULL) OR (target_comment_id IS NULL)))
);


ALTER TABLE public.comment_ OWNER TO postgres;

--
-- Name: pc_reaction; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pc_reaction (
    id integer NOT NULL,
    reactor_user_id integer NOT NULL,
    target_post_id integer,
    target_comment_id integer,
    reaction_code_point integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT reaction_either_in_post_or_comment CHECK (((target_post_id IS NULL) OR (target_comment_id IS NULL)))
);


ALTER TABLE public.pc_reaction OWNER TO postgres;

--
-- Name: CommentView; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public."CommentView" AS
 SELECT i9l_user.id AS owner_user_id,
    i9l_user.username AS owner_username,
    i9l_user.profile_pic_url AS owner_profile_pic_url,
    cm.id AS comment_id,
    cm.comment_text,
    cm.attachment_url,
    (count(DISTINCT any_reaction.id))::integer AS reactions_count,
    (count(DISTINCT cm_on_cm.id))::integer AS comments_count,
    certain_reaction.reactor_user_id,
    certain_reaction.reaction_code_point,
    cm.target_post_id,
    cm.target_comment_id,
    cm.created_at
   FROM ((((public.comment_ cm
     JOIN public.i9l_user ON ((i9l_user.id = cm.commenter_user_id)))
     LEFT JOIN public.pc_reaction any_reaction ON ((any_reaction.target_comment_id = cm.id)))
     LEFT JOIN public.comment_ cm_on_cm ON ((cm_on_cm.target_comment_id = cm.id)))
     LEFT JOIN public.pc_reaction certain_reaction ON ((certain_reaction.target_comment_id = cm.id)))
  GROUP BY i9l_user.id, i9l_user.username, i9l_user.profile_pic_url, cm.id, cm.comment_text, cm.attachment_url, certain_reaction.reactor_user_id, certain_reaction.reaction_code_point, cm.target_post_id, cm.target_comment_id, cm.created_at;


ALTER VIEW public."CommentView" OWNER TO postgres;

--
-- Name: post; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.post (
    id integer NOT NULL,
    user_id integer NOT NULL,
    media_urls text[] NOT NULL,
    description text,
    type text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.post OWNER TO postgres;

--
-- Name: repost; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.repost (
    id integer NOT NULL,
    reposter_user_id integer NOT NULL,
    post_id integer NOT NULL
);


ALTER TABLE public.repost OWNER TO postgres;

--
-- Name: saved_post; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.saved_post (
    id integer NOT NULL,
    saver_user_id integer NOT NULL,
    post_id integer NOT NULL
);


ALTER TABLE public.saved_post OWNER TO postgres;

--
-- Name: PostView; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public."PostView" AS
 SELECT i9l_user.id AS owner_user_id,
    i9l_user.username AS owner_username,
    i9l_user.profile_pic_url AS owner_profile_pic_url,
    post.id AS post_id,
    post.type,
    post.media_urls,
    post.description,
    (count(DISTINCT any_reaction.id))::integer AS reactions_count,
    (count(DISTINCT any_comment.id))::integer AS comments_count,
    (count(DISTINCT any_repost.id))::integer AS reposts_count,
    (count(DISTINCT any_saved_post.id))::integer AS saves_count,
    certain_reaction.reactor_user_id,
    certain_reaction.reaction_code_point,
    certain_repost.reposter_user_id,
    certain_saved_post.saver_user_id,
    post.created_at
   FROM ((((((((public.post
     JOIN public.i9l_user ON ((i9l_user.id = post.user_id)))
     LEFT JOIN public.pc_reaction any_reaction ON ((any_reaction.target_post_id = post.id)))
     LEFT JOIN public.comment_ any_comment ON ((any_comment.target_post_id = post.id)))
     LEFT JOIN public.repost any_repost ON ((any_repost.post_id = post.id)))
     LEFT JOIN public.saved_post any_saved_post ON ((any_saved_post.post_id = post.id)))
     LEFT JOIN public.pc_reaction certain_reaction ON ((certain_reaction.target_post_id = post.id)))
     LEFT JOIN public.repost certain_repost ON ((certain_repost.post_id = post.id)))
     LEFT JOIN public.saved_post certain_saved_post ON ((certain_saved_post.post_id = post.id)))
  GROUP BY i9l_user.id, i9l_user.username, i9l_user.profile_pic_url, post.id, post.type, post.media_urls, post.description, certain_reaction.reactor_user_id, certain_reaction.reaction_code_point, certain_repost.reposter_user_id, certain_saved_post.saver_user_id, post.created_at;


ALTER VIEW public."PostView" OWNER TO postgres;

--
-- Name: blocked_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.blocked_user (
    id integer NOT NULL,
    blocking_user_id integer NOT NULL,
    blocked_user_id integer NOT NULL,
    blocked_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.blocked_user OWNER TO postgres;

--
-- Name: blocked_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.blocked_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.blocked_user_id_seq OWNER TO postgres;

--
-- Name: blocked_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.blocked_user_id_seq OWNED BY public.blocked_user.id;


--
-- Name: comment_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.comment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.comment_id_seq OWNER TO postgres;

--
-- Name: comment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.comment_id_seq OWNED BY public.comment_.id;


--
-- Name: conversation; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.conversation (
    id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    initiator_user_id integer NOT NULL,
    with_user_id integer NOT NULL
);


ALTER TABLE public.conversation OWNER TO postgres;

--
-- Name: conversation_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.conversation_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.conversation_id_seq OWNER TO postgres;

--
-- Name: conversation_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.conversation_id_seq OWNED BY public.conversation.id;


--
-- Name: follow; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.follow (
    id integer NOT NULL,
    follower_user_id integer NOT NULL,
    followee_user_id integer NOT NULL,
    follow_on timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT no_self_follow CHECK ((follower_user_id <> followee_user_id))
);


ALTER TABLE public.follow OWNER TO postgres;

--
-- Name: follow_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.follow_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.follow_id_seq OWNER TO postgres;

--
-- Name: follow_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.follow_id_seq OWNED BY public.follow.id;


--
-- Name: i9l_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.i9l_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.i9l_user_id_seq OWNER TO postgres;

--
-- Name: i9l_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.i9l_user_id_seq OWNED BY public.i9l_user.id;


--
-- Name: message_deletion_log; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.message_deletion_log (
    id integer NOT NULL,
    deleter_user_id integer NOT NULL,
    message_id integer NOT NULL,
    deleted_for character varying NOT NULL,
    CONSTRAINT message_deletion_log_deleted_for_check CHECK (((deleted_for)::text = ANY (ARRAY['me'::text, 'everyone'::text])))
);


ALTER TABLE public.message_deletion_log OWNER TO postgres;

--
-- Name: message_deletion_log_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.message_deletion_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.message_deletion_log_id_seq OWNER TO postgres;

--
-- Name: message_deletion_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.message_deletion_log_id_seq OWNED BY public.message_deletion_log.id;


--
-- Name: message_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.message_id_seq OWNER TO postgres;

--
-- Name: message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.message_id_seq OWNED BY public.message_.id;


--
-- Name: message_reaction_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.message_reaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.message_reaction_id_seq OWNER TO postgres;

--
-- Name: message_reaction_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.message_reaction_id_seq OWNED BY public.message_reaction.id;


--
-- Name: notification; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notification (
    id integer NOT NULL,
    type character varying(255) NOT NULL,
    is_read boolean DEFAULT false,
    sender_user_id integer NOT NULL,
    receiver_user_id integer NOT NULL,
    via_post_id integer,
    via_comment_id integer,
    comment_created_id integer,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    reaction_code_point integer
);


ALTER TABLE public.notification OWNER TO postgres;

--
-- Name: notification_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.notification_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.notification_id_seq OWNER TO postgres;

--
-- Name: notification_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.notification_id_seq OWNED BY public.notification.id;


--
-- Name: ongoing_registration; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ongoing_registration (
    sid character varying NOT NULL,
    sess json NOT NULL,
    expire timestamp(6) without time zone NOT NULL
);


ALTER TABLE public.ongoing_registration OWNER TO postgres;

--
-- Name: pc_hashtag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pc_hashtag (
    id integer NOT NULL,
    post_id integer,
    comment_id integer,
    hashtag_name character varying(255) NOT NULL,
    CONSTRAINT hashtag_either_in_post_or_comment CHECK (((post_id IS NULL) OR (comment_id IS NULL)))
);


ALTER TABLE public.pc_hashtag OWNER TO postgres;

--
-- Name: pc_hashtag_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.pc_hashtag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pc_hashtag_id_seq OWNER TO postgres;

--
-- Name: pc_hashtag_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.pc_hashtag_id_seq OWNED BY public.pc_hashtag.id;


--
-- Name: pc_mention; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pc_mention (
    id integer NOT NULL,
    post_id integer,
    comment_id integer,
    user_id integer NOT NULL,
    CONSTRAINT mention_either_in_post_or_comment CHECK (((post_id IS NULL) OR (comment_id IS NULL)))
);


ALTER TABLE public.pc_mention OWNER TO postgres;

--
-- Name: pc_mention_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.pc_mention_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pc_mention_id_seq OWNER TO postgres;

--
-- Name: pc_mention_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.pc_mention_id_seq OWNED BY public.pc_mention.id;


--
-- Name: pc_reaction_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.pc_reaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pc_reaction_id_seq OWNER TO postgres;

--
-- Name: pc_reaction_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.pc_reaction_id_seq OWNED BY public.pc_reaction.id;


--
-- Name: post_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.post_id_seq OWNER TO postgres;

--
-- Name: post_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.post_id_seq OWNED BY public.post.id;


--
-- Name: reported_message; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.reported_message (
    id integer NOT NULL,
    reporting_user_id integer NOT NULL,
    reported_user_id integer NOT NULL,
    message_id integer NOT NULL,
    reason text NOT NULL,
    reported_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.reported_message OWNER TO postgres;

--
-- Name: reported_message_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.reported_message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.reported_message_id_seq OWNER TO postgres;

--
-- Name: reported_message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.reported_message_id_seq OWNED BY public.reported_message.id;


--
-- Name: repost_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.repost_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.repost_id_seq OWNER TO postgres;

--
-- Name: repost_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.repost_id_seq OWNED BY public.repost.id;


--
-- Name: saved_post_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.saved_post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.saved_post_id_seq OWNER TO postgres;

--
-- Name: saved_post_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.saved_post_id_seq OWNED BY public.saved_post.id;


--
-- Name: user_conversation; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_conversation (
    id integer NOT NULL,
    user_id integer NOT NULL,
    conversation_id integer NOT NULL,
    unread_messages_count integer DEFAULT 0,
    notification_mode text DEFAULT 'enabled'::text NOT NULL,
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    partner_user_id integer NOT NULL,
    CONSTRAINT "UserConversation_notification_mode_check" CHECK ((notification_mode = ANY (ARRAY['enabled'::text, 'mute'::text])))
);


ALTER TABLE public.user_conversation OWNER TO postgres;

--
-- Name: user_conversation_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_conversation_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_conversation_id_seq OWNER TO postgres;

--
-- Name: user_conversation_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_conversation_id_seq OWNED BY public.user_conversation.id;


--
-- Name: blocked_user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.blocked_user ALTER COLUMN id SET DEFAULT nextval('public.blocked_user_id_seq'::regclass);


--
-- Name: comment_ id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_ ALTER COLUMN id SET DEFAULT nextval('public.comment_id_seq'::regclass);


--
-- Name: conversation id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.conversation ALTER COLUMN id SET DEFAULT nextval('public.conversation_id_seq'::regclass);


--
-- Name: follow id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.follow ALTER COLUMN id SET DEFAULT nextval('public.follow_id_seq'::regclass);


--
-- Name: i9l_user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.i9l_user ALTER COLUMN id SET DEFAULT nextval('public.i9l_user_id_seq'::regclass);


--
-- Name: message_ id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_ ALTER COLUMN id SET DEFAULT nextval('public.message_id_seq'::regclass);


--
-- Name: message_deletion_log id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_deletion_log ALTER COLUMN id SET DEFAULT nextval('public.message_deletion_log_id_seq'::regclass);


--
-- Name: message_reaction id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_reaction ALTER COLUMN id SET DEFAULT nextval('public.message_reaction_id_seq'::regclass);


--
-- Name: notification id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification ALTER COLUMN id SET DEFAULT nextval('public.notification_id_seq'::regclass);


--
-- Name: pc_hashtag id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_hashtag ALTER COLUMN id SET DEFAULT nextval('public.pc_hashtag_id_seq'::regclass);


--
-- Name: pc_mention id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_mention ALTER COLUMN id SET DEFAULT nextval('public.pc_mention_id_seq'::regclass);


--
-- Name: pc_reaction id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_reaction ALTER COLUMN id SET DEFAULT nextval('public.pc_reaction_id_seq'::regclass);


--
-- Name: post id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.post ALTER COLUMN id SET DEFAULT nextval('public.post_id_seq'::regclass);


--
-- Name: reported_message id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reported_message ALTER COLUMN id SET DEFAULT nextval('public.reported_message_id_seq'::regclass);


--
-- Name: repost id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repost ALTER COLUMN id SET DEFAULT nextval('public.repost_id_seq'::regclass);


--
-- Name: saved_post id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saved_post ALTER COLUMN id SET DEFAULT nextval('public.saved_post_id_seq'::regclass);


--
-- Name: user_conversation id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_conversation ALTER COLUMN id SET DEFAULT nextval('public.user_conversation_id_seq'::regclass);


--
-- Data for Name: blocked_user; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.blocked_user (id, blocking_user_id, blocked_user_id, blocked_at) FROM stdin;
\.


--
-- Data for Name: comment_; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.comment_ (id, comment_text, commenter_user_id, attachment_url, target_post_id, target_comment_id, created_at) FROM stdin;
1	This is a comment on this post from @johnny.	10		8	\N	2024-07-15 12:13:58.721514
3	This is a comment on this post from @starlight.	13		8	\N	2024-07-15 12:17:04.994458
4	This is a comment on this post from @itz_butcher.	12		9	\N	2024-07-15 12:32:25.849147
5	This is a comment on this post from @kendrick.	11		9	\N	2024-07-15 12:32:25.853729
6	This is a comment on this post from @kendrick.	11		10	\N	2024-07-15 12:32:25.988137
7	This is a comment on this post from @johnny.	10		10	\N	2024-07-15 12:32:26.234091
8	This is a comment on this post from @starlight.	13		4	\N	2024-07-15 12:32:26.306577
9	This is a comment on this post from @itz_butcher.	12		4	\N	2024-07-15 12:32:26.310416
11	This is a comment on this post from @kendrick.	11		5	\N	2024-07-15 12:32:26.327614
12	This is a comment on this post from @starlight.	13		5	\N	2024-07-15 12:32:26.335346
13	This is a comment on this post from @johnny.	10		5	\N	2024-07-15 12:32:26.336911
14	This is a comment on this post from @kendrick.	11		4	\N	2024-07-15 12:32:26.340869
15	This is a comment on this post from @itz_butcher.	12		10	\N	2024-07-15 12:32:26.35482
18	This is a reply to this comment from @kendrick.	11		\N	13	2024-07-15 13:08:20.948356
19	This is a reply to this comment from @johnny.	10		\N	14	2024-07-15 13:08:21.078347
20	This is a reply to this comment from @kendrick.	11		\N	11	2024-07-15 13:08:21.65685
21	This is a reply to this comment from @starlight.	13		\N	5	2024-07-15 13:08:21.657533
23	This is a reply to this comment from @johnny.	10		\N	6	2024-07-15 13:08:21.704893
24	This is a reply to this comment from @starlight.	13		\N	14	2024-07-15 13:08:21.723639
26	This is a reply to this comment from @johnny.	10		\N	7	2024-07-15 13:08:21.730694
27	This is a reply to this comment from @itz_butcher.	12		\N	6	2024-07-15 13:08:21.737682
28	This is a reply to this comment from @starlight.	13		\N	5	2024-07-15 13:08:21.748516
29	This is a reply to this comment from @itz_butcher.	12		\N	7	2024-07-15 13:08:21.745679
30	This is a reply to this comment from @starlight.	13		\N	13	2024-07-15 13:08:21.752653
31	This is a reply to this comment from @itz_butcher.	12		\N	1	2024-07-15 13:08:21.753414
32	This is a reply to this comment from @kendrick.	11		\N	15	2024-07-15 13:08:21.80434
33	This is a reply to this comment from @kendrick.	11		\N	4	2024-07-15 13:08:21.839795
34	This is a reply to this comment from @starlight.	13		\N	4	2024-07-15 13:08:21.840976
35	This is a reply to this comment from @starlight.	13		\N	15	2024-07-15 13:08:21.842159
36	This is a reply to this comment from @johnny.	10		\N	4	2024-07-15 13:08:21.843322
37	This is a reply to this comment from @johnny.	10		\N	9	2024-07-15 13:08:21.844372
38	This is a reply to this comment from @johnny.	10		\N	15	2024-07-15 13:08:21.845693
39	This is a reply to this comment from @itz_butcher.	12		\N	12	2024-07-15 13:08:21.846941
40	This is a reply to this comment from @itz_butcher.	12		\N	3	2024-07-15 13:08:21.848118
41	This is a reply to this comment from @kendrick.	11		\N	8	2024-07-15 13:08:21.849203
42	This is a reply to this comment from @kendrick.	11		\N	3	2024-07-15 13:08:21.916378
43	This is a reply to this comment from @johnny.	10		\N	8	2024-07-15 13:08:21.960947
44	This is a reply to this comment from @johnny.	10		\N	3	2024-07-15 13:08:21.962895
45	This is a reply to this comment from @johnny.	10		\N	12	2024-07-15 13:08:21.964706
\.


--
-- Data for Name: conversation; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.conversation (id, created_at, initiator_user_id, with_user_id) FROM stdin;
1	2024-07-15 21:18:32.031916	10	12
\.


--
-- Data for Name: follow; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.follow (id, follower_user_id, followee_user_id, follow_on) FROM stdin;
37	10	11	2024-07-13 23:52:58.602132
38	10	13	2024-07-13 23:53:22.779341
39	12	11	2024-07-13 23:54:08.396625
40	12	13	2024-07-13 23:54:23.955964
41	12	10	2024-07-13 23:54:47.444132
78	13	10	2024-07-14 16:51:51.236682
79	11	10	2024-07-14 16:52:22.864408
82	11	13	2024-07-14 17:36:04.614918
83	13	11	2024-07-14 17:38:56.101537
\.


--
-- Data for Name: i9l_user; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.i9l_user (id, email, username, password, name, birthday, bio, profile_pic_url, connection_status, last_active, acc_deleted, cover_pic_url) FROM stdin;
11	annak@gmail.com	kendrick	$2b$10$HgR1Onh76X32zHWFryZG9Oizc.cND9dyyTdZts.IFc.8F43BqVzrS	Anna Kendrick	2000-12-07	#musicIsLife		online	\N	f	
12	butcher@gmail.com	itz_butcher	$2b$10$.t6L5VxQkAf8NtQXVSKepeHz4Z1JC/xnGit4GKs/jMxYk0zgWT8iS	William Butcher	2000-11-07	Alright then, love		online	\N	f	
13	annie_star@gmail.com	starlight	$2b$10$FxIqb8LpmnupXJDkaBURwuv3zIGPgRFSde1EPNAyJX8Fe9igV6YQC	Annie January	2000-01-07	I'm a good Supe! Pls, don't hurt me.		online	\N	f	
10	johnny@gmail.com	johnny	$2b$10$RrSgFcssPSMTFb6SW.CeeOdSfesv66l6ipDgUZjVBjdonlJj1BM6W	Samuel Ayomide	2000-12-07	#nerdIsLife		offline	2024-08-26 07:28:37.626	f	
\.


--
-- Data for Name: message_; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.message_ (id, sender_user_id, conversation_id, msg_content, created_at, delivery_status, reply_to_id) FROM stdin;
5	10	1	{"type": "text", "props": {"textContent": "It's really boring here."}}	2024-07-15 21:54:24.101866	sent	\N
6	12	1	{"type": "text", "props": {"textContent": "Me too, love."}}	2024-07-15 21:54:24.049857	sent	\N
7	10	1	{"type": "text", "props": {"textContent": "How's it going over there?"}}	2024-07-15 21:54:24.108832	sent	\N
1	10	1	{"type": "text", "props": {"textContent": "Hi! How're you?"}}	2024-07-15 21:18:32.031916	read	\N
4	10	1	{"type": "text", "props": {"textContent": "I've missed you, man."}}	2024-07-15 21:54:24.045331	delivered	\N
3	12	1	{"type": "text", "props": {"textContent": "Heeeyy! I'm fine!"}}	2024-07-15 21:54:24.043517	read	\N
\.


--
-- Data for Name: message_deletion_log; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.message_deletion_log (id, deleter_user_id, message_id, deleted_for) FROM stdin;
1	12	3	me
2	12	3	me
\.


--
-- Data for Name: message_reaction; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.message_reaction (id, message_id, reactor_user_id, reaction_code_point) FROM stdin;
2	1	12	129392
3	3	10	129392
\.


--
-- Data for Name: notification; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.notification (id, type, is_read, sender_user_id, receiver_user_id, via_post_id, via_comment_id, comment_created_id, created_at, reaction_code_point) FROM stdin;
1	follow	f	10	12	\N	\N	\N	2024-07-12 23:50:26.139837	\N
2	follow	f	10	12	\N	\N	\N	2024-07-12 23:57:48.133812	\N
3	follow	f	10	12	\N	\N	\N	2024-07-13 00:00:44.96956	\N
4	follow	f	10	12	\N	\N	\N	2024-07-13 20:58:17.771663	\N
5	follow	f	10	12	\N	\N	\N	2024-07-13 21:09:44.274479	\N
6	follow	f	10	12	\N	\N	\N	2024-07-13 21:15:16.5168	\N
7	follow	f	10	12	\N	\N	\N	2024-07-13 21:17:53.768451	\N
8	follow	f	10	12	\N	\N	\N	2024-07-13 21:27:39.707541	\N
9	follow	f	10	12	\N	\N	\N	2024-07-13 22:16:59.039484	\N
10	follow	f	10	12	\N	\N	\N	2024-07-13 22:21:07.815159	\N
11	follow	f	10	12	\N	\N	\N	2024-07-13 22:28:23.720249	\N
12	follow	f	10	12	\N	\N	\N	2024-07-13 22:31:40.431423	\N
13	follow	f	10	12	\N	\N	\N	2024-07-13 22:38:47.592551	\N
14	follow	f	10	12	\N	\N	\N	2024-07-13 22:41:14.729691	\N
15	follow	f	10	12	\N	\N	\N	2024-07-13 22:41:50.504706	\N
16	follow	f	10	12	\N	\N	\N	2024-07-13 23:04:56.11296	\N
17	follow	f	10	12	\N	\N	\N	2024-07-13 23:08:56.666052	\N
18	follow	f	10	12	\N	\N	\N	2024-07-13 23:10:02.98658	\N
19	follow	f	10	12	\N	\N	\N	2024-07-13 23:11:14.396798	\N
20	follow	f	10	12	\N	\N	\N	2024-07-13 23:13:50.959412	\N
21	follow	f	10	12	\N	\N	\N	2024-07-13 23:17:34.017002	\N
22	follow	f	10	12	\N	\N	\N	2024-07-13 23:19:06.053373	\N
23	follow	f	10	12	\N	\N	\N	2024-07-13 23:24:27.237355	\N
24	follow	f	10	12	\N	\N	\N	2024-07-13 23:27:32.638363	\N
25	follow	f	10	12	\N	\N	\N	2024-07-13 23:34:20.672912	\N
26	follow	f	10	12	\N	\N	\N	2024-07-13 23:35:00.568125	\N
27	follow	f	10	12	\N	\N	\N	2024-07-13 23:35:56.805973	\N
28	follow	f	10	12	\N	\N	\N	2024-07-13 23:37:35.381475	\N
29	follow	f	10	12	\N	\N	\N	2024-07-13 23:38:53.242301	\N
30	follow	f	10	12	\N	\N	\N	2024-07-13 23:40:38.975742	\N
31	follow	f	10	12	\N	\N	\N	2024-07-13 23:42:30.404232	\N
32	follow	f	10	12	\N	\N	\N	2024-07-13 23:46:34.850464	\N
33	follow	f	10	12	\N	\N	\N	2024-07-13 23:49:11.510618	\N
34	follow	f	10	12	\N	\N	\N	2024-07-13 23:52:38.830575	\N
35	follow	f	10	11	\N	\N	\N	2024-07-13 23:52:58.602132	\N
36	follow	f	10	13	\N	\N	\N	2024-07-13 23:53:22.779341	\N
37	follow	f	12	11	\N	\N	\N	2024-07-13 23:54:08.396625	\N
38	follow	f	12	13	\N	\N	\N	2024-07-13 23:54:23.955964	\N
39	follow	f	12	10	\N	\N	\N	2024-07-13 23:54:47.444132	\N
40	follow	f	10	12	\N	\N	\N	2024-07-13 23:56:22.116429	\N
73	follow	f	13	10	\N	\N	\N	2024-07-14 16:51:51.236682	\N
74	follow	f	11	10	\N	\N	\N	2024-07-14 16:52:22.864408	\N
75	follow	f	10	12	\N	\N	\N	2024-07-14 16:53:27.485798	\N
76	follow	f	11	13	\N	\N	\N	2024-07-14 17:36:04.614918	\N
77	follow	f	13	11	\N	\N	\N	2024-07-14 17:38:56.101537	\N
78	mention_in_post	f	10	12	4	\N	\N	2024-07-14 21:04:14.940727	\N
79	reaction_to_post	f	12	10	4	\N	\N	2024-07-15 09:12:12.293995	129315
80	reaction_to_post	f	11	10	4	\N	\N	2024-07-15 09:15:17.903499	129315
81	reaction_to_post	f	13	10	4	\N	\N	2024-07-15 09:15:58.840223	129315
82	reaction_to_post	f	10	12	5	\N	\N	2024-07-15 09:17:13.849673	129315
83	reaction_to_post	f	12	12	5	\N	\N	2024-07-15 09:17:32.83413	129315
84	reaction_to_post	f	13	12	5	\N	\N	2024-07-15 09:17:49.484283	129315
85	reaction_to_post	f	11	12	5	\N	\N	2024-07-15 09:17:59.678161	129315
86	reaction_to_post	f	10	11	8	\N	\N	2024-07-15 09:19:45.797318	128536
87	reaction_to_post	f	13	11	8	\N	\N	2024-07-15 09:20:10.058672	128536
88	reaction_to_post	f	12	13	10	\N	\N	2024-07-15 09:39:11.801503	128514
89	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 09:42:39.462862	128514
90	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 09:48:04.353943	128514
91	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 09:50:32.935136	128514
92	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 09:52:41.967722	128514
93	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 09:56:01.51688	128514
94	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 09:58:02.929425	128514
95	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 09:58:15.00754	128514
96	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 09:58:35.883797	128514
97	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 10:00:34.588546	128514
98	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 10:02:18.483394	128514
99	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 10:33:35.876517	128514
100	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 11:30:31.98657	128514
101	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 11:34:29.460505	128514
102	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 12:08:51.707568	128514
103	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 12:13:58.016895	128514
104	comment_on_post	f	10	11	8	\N	1	2024-07-15 12:13:58.721514	\N
105	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 12:14:42.006721	128514
239	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 14:14:12.215078	128514
107	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 12:17:04.289659	128514
108	comment_on_post	f	13	11	8	\N	3	2024-07-15 12:17:04.994458	\N
109	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 12:32:24.891424	128514
110	comment_on_post	f	12	11	9	\N	4	2024-07-15 12:32:25.849147	\N
111	comment_on_post	f	11	11	9	\N	5	2024-07-15 12:32:25.853729	\N
112	comment_on_post	f	11	13	10	\N	6	2024-07-15 12:32:25.988137	\N
113	comment_on_post	f	10	13	10	\N	7	2024-07-15 12:32:26.234091	\N
114	comment_on_post	f	13	10	4	\N	8	2024-07-15 12:32:26.306577	\N
115	comment_on_post	f	12	10	4	\N	9	2024-07-15 12:32:26.310416	\N
117	comment_on_post	f	11	12	5	\N	11	2024-07-15 12:32:26.327614	\N
118	comment_on_post	f	13	12	5	\N	12	2024-07-15 12:32:26.335346	\N
119	comment_on_post	f	10	12	5	\N	13	2024-07-15 12:32:26.336911	\N
120	comment_on_post	f	11	10	4	\N	14	2024-07-15 12:32:26.340869	\N
121	comment_on_post	f	12	13	10	\N	15	2024-07-15 12:32:26.35482	\N
122	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 12:56:44.80052	128514
242	reaction_to_comment	f	11	10	\N	13	\N	2024-07-15 14:14:13.145348	129463
244	reaction_to_comment	f	13	11	\N	5	\N	2024-07-15 14:14:13.358042	128148
246	reaction_to_comment	f	11	12	\N	15	\N	2024-07-15 14:14:13.76665	127919
248	reaction_to_comment	f	13	12	\N	4	\N	2024-07-15 14:14:13.82018	127919
253	reaction_to_comment	f	12	11	\N	4	\N	2024-07-15 14:14:13.847896	129463
258	reaction_to_comment	f	10	12	\N	4	\N	2024-07-15 14:14:13.933688	128148
261	reaction_to_comment	f	12	13	\N	12	\N	2024-07-15 14:14:13.967355	127919
268	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 14:25:02.256061	128514
271	reaction_to_comment	f	11	12	\N	4	\N	2024-07-15 14:25:02.765869	129463
288	follow	f	10	12	\N	\N	\N	2024-07-15 15:25:07.617782	\N
243	reaction_to_comment	f	13	10	\N	13	\N	2024-07-15 14:14:13.149765	128536
245	reaction_to_comment	f	13	11	\N	14	\N	2024-07-15 14:14:13.431344	128530
247	reaction_to_comment	f	11	12	\N	4	\N	2024-07-15 14:14:13.767665	129463
252	reaction_to_comment	f	13	12	\N	15	\N	2024-07-15 14:14:13.867118	128536
254	reaction_to_comment	f	12	10	\N	1	\N	2024-07-15 14:14:13.853306	128530
262	reaction_to_comment	f	12	13	\N	3	\N	2024-07-15 14:14:13.968264	129463
267	reaction_to_comment	f	10	13	\N	12	\N	2024-07-15 14:14:13.971891	128530
280	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 14:58:18.254411	128514
283	reaction_to_comment	f	11	12	\N	4	\N	2024-07-15 14:58:18.952893	129463
289	follow	f	10	12	\N	\N	\N	2024-07-15 15:35:32.650289	\N
249	reaction_to_comment	f	10	10	\N	7	\N	2024-07-15 14:14:13.829536	128148
255	reaction_to_comment	f	12	10	\N	7	\N	2024-07-15 14:14:13.86766	128148
259	reaction_to_comment	f	10	12	\N	9	\N	2024-07-15 14:14:13.934437	128148
266	reaction_to_comment	f	10	13	\N	3	\N	2024-07-15 14:14:13.971402	128148
272	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 14:48:53.331937	128514
275	reaction_to_comment	f	11	12	\N	4	\N	2024-07-15 14:48:53.986142	129463
290	follow	f	10	12	\N	\N	\N	2024-07-15 15:37:49.135239	\N
250	reaction_to_comment	f	10	11	\N	14	\N	2024-07-15 14:14:13.833029	128536
256	reaction_to_comment	f	11	11	\N	11	\N	2024-07-15 14:14:13.872053	128148
260	reaction_to_comment	f	10	12	\N	15	\N	2024-07-15 14:14:13.966713	128530
264	reaction_to_comment	f	11	13	\N	3	\N	2024-07-15 14:14:13.969804	128536
284	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 14:59:39.294983	128514
251	reaction_to_comment	f	12	11	\N	6	\N	2024-07-15 14:14:13.838701	127919
153	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 13:02:36.363584	128514
257	reaction_to_comment	f	10	11	\N	6	\N	2024-07-15 14:14:13.878731	127919
263	reaction_to_comment	f	11	13	\N	8	\N	2024-07-15 14:14:13.969137	127919
265	reaction_to_comment	f	10	13	\N	8	\N	2024-07-15 14:14:13.970606	128148
276	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 14:56:26.679009	128514
279	reaction_to_comment	f	11	12	\N	4	\N	2024-07-15 14:56:27.461828	129463
287	reaction_to_comment	f	11	12	\N	4	\N	2024-07-15 14:59:40.003424	129463
182	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 13:08:20.255595	128514
185	comment_on_comment	f	11	10	\N	13	18	2024-07-15 13:08:20.948356	\N
186	comment_on_comment	f	10	11	\N	14	19	2024-07-15 13:08:21.078347	\N
187	comment_on_comment	f	11	11	\N	11	20	2024-07-15 13:08:21.65685	\N
188	comment_on_comment	f	13	11	\N	5	21	2024-07-15 13:08:21.657533	\N
190	comment_on_comment	f	10	11	\N	6	23	2024-07-15 13:08:21.704893	\N
192	comment_on_comment	f	13	10	\N	13	30	2024-07-15 13:08:21.752653	\N
193	comment_on_comment	f	13	11	\N	5	28	2024-07-15 13:08:21.748516	\N
194	comment_on_comment	f	12	11	\N	6	27	2024-07-15 13:08:21.737682	\N
195	comment_on_comment	f	13	11	\N	14	24	2024-07-15 13:08:21.723639	\N
196	comment_on_comment	f	12	10	\N	1	31	2024-07-15 13:08:21.753414	\N
197	comment_on_comment	f	12	10	\N	7	29	2024-07-15 13:08:21.745679	\N
198	comment_on_comment	f	10	10	\N	7	26	2024-07-15 13:08:21.730694	\N
199	comment_on_comment	f	11	12	\N	15	32	2024-07-15 13:08:21.80434	\N
200	comment_on_comment	f	11	12	\N	4	33	2024-07-15 13:08:21.839795	\N
201	comment_on_comment	f	13	12	\N	4	34	2024-07-15 13:08:21.840976	\N
202	comment_on_comment	f	13	12	\N	15	35	2024-07-15 13:08:21.842159	\N
203	comment_on_comment	f	10	12	\N	4	36	2024-07-15 13:08:21.843322	\N
204	comment_on_comment	f	10	12	\N	9	37	2024-07-15 13:08:21.844372	\N
205	comment_on_comment	f	10	12	\N	15	38	2024-07-15 13:08:21.845693	\N
206	comment_on_comment	f	12	13	\N	12	39	2024-07-15 13:08:21.846941	\N
207	comment_on_comment	f	12	13	\N	3	40	2024-07-15 13:08:21.848118	\N
208	comment_on_comment	f	11	13	\N	8	41	2024-07-15 13:08:21.849203	\N
209	comment_on_comment	f	11	13	\N	3	42	2024-07-15 13:08:21.916378	\N
210	comment_on_comment	f	10	13	\N	8	43	2024-07-15 13:08:21.960947	\N
211	comment_on_comment	f	10	13	\N	3	44	2024-07-15 13:08:21.962895	\N
212	comment_on_comment	f	10	13	\N	12	45	2024-07-15 13:08:21.964706	\N
213	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 13:13:26.259693	128514
216	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 13:42:29.66134	128514
219	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 14:01:14.855458	128514
222	reaction_to_comment	f	11	10	\N	13	\N	2024-07-15 14:01:15.866319	129463
223	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 14:06:06.256163	128514
226	reaction_to_comment	f	11	10	\N	13	\N	2024-07-15 14:06:07.179683	129463
227	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 14:07:59.287745	128514
230	reaction_to_comment	f	11	10	\N	13	\N	2024-07-15 14:08:00.377297	129463
231	reaction_to_post	f	11	13	10	\N	\N	2024-07-15 14:10:07.255311	128514
234	reaction_to_comment	f	11	10	\N	13	\N	2024-07-15 14:10:08.231689	129463
235	reaction_to_comment	f	13	10	\N	13	\N	2024-07-15 14:10:08.236102	128536
236	reaction_to_comment	f	13	11	\N	5	\N	2024-07-15 14:10:08.43673	128148
237	reaction_to_comment	f	13	11	\N	14	\N	2024-07-15 14:10:08.482907	128530
238	reaction_to_comment	f	11	12	\N	15	\N	2024-07-15 14:10:08.786926	127919
\.


--
-- Data for Name: ongoing_registration; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.ongoing_registration (sid, sess, expire) FROM stdin;
\.


--
-- Data for Name: pc_hashtag; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pc_hashtag (id, post_id, comment_id, hashtag_name) FROM stdin;
2	4	\N	willy
\.


--
-- Data for Name: pc_mention; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pc_mention (id, post_id, comment_id, user_id) FROM stdin;
3	4	\N	12
4	\N	1	10
6	\N	3	13
7	\N	4	12
8	\N	5	11
9	\N	6	11
10	\N	7	10
11	\N	8	13
12	\N	9	12
14	\N	11	11
15	\N	12	13
16	\N	13	10
17	\N	14	11
18	\N	15	12
79	\N	18	11
80	\N	19	10
81	\N	20	11
82	\N	21	13
84	\N	23	10
86	\N	30	13
88	\N	28	13
87	\N	27	12
89	\N	24	13
90	\N	29	12
91	\N	31	12
92	\N	26	10
93	\N	32	11
94	\N	33	11
95	\N	34	13
96	\N	35	13
97	\N	36	10
98	\N	37	10
99	\N	38	10
100	\N	39	12
101	\N	40	12
102	\N	41	11
103	\N	42	11
104	\N	43	10
105	\N	44	10
106	\N	45	10
\.


--
-- Data for Name: pc_reaction; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pc_reaction (id, reactor_user_id, target_post_id, target_comment_id, reaction_code_point, created_at) FROM stdin;
1	12	4	\N	129315	2024-07-15 09:12:12.293995
2	11	4	\N	129315	2024-07-15 09:15:17.903499
3	13	4	\N	129315	2024-07-15 09:15:58.840223
4	10	5	\N	129315	2024-07-15 09:17:13.849673
5	12	5	\N	129315	2024-07-15 09:17:32.83413
6	13	5	\N	129315	2024-07-15 09:17:49.484283
7	11	5	\N	129315	2024-07-15 09:17:59.678161
8	10	8	\N	128536	2024-07-15 09:19:45.797318
9	13	8	\N	128536	2024-07-15 09:20:10.058672
10	12	10	\N	128514	2024-07-15 09:39:11.801503
12	11	\N	13	129463	2024-07-15 14:14:13.145348
13	13	\N	13	128536	2024-07-15 14:14:13.149765
14	13	\N	5	128148	2024-07-15 14:14:13.358042
15	13	\N	14	128530	2024-07-15 14:14:13.431344
16	11	\N	15	127919	2024-07-15 14:14:13.76665
18	13	\N	4	127919	2024-07-15 14:14:13.82018
19	10	\N	7	128148	2024-07-15 14:14:13.829536
20	10	\N	14	128536	2024-07-15 14:14:13.833029
21	12	\N	6	127919	2024-07-15 14:14:13.838701
22	12	\N	1	128530	2024-07-15 14:14:13.853306
23	12	\N	4	129463	2024-07-15 14:14:13.847896
24	13	\N	15	128536	2024-07-15 14:14:13.867118
25	12	\N	7	128148	2024-07-15 14:14:13.86766
26	11	\N	11	128148	2024-07-15 14:14:13.872053
27	10	\N	6	127919	2024-07-15 14:14:13.878731
28	10	\N	4	128148	2024-07-15 14:14:13.933688
29	10	\N	9	128148	2024-07-15 14:14:13.934437
30	10	\N	15	128530	2024-07-15 14:14:13.966713
31	12	\N	12	127919	2024-07-15 14:14:13.967355
32	12	\N	3	129463	2024-07-15 14:14:13.968264
33	11	\N	8	127919	2024-07-15 14:14:13.969137
34	11	\N	3	128536	2024-07-15 14:14:13.969804
35	10	\N	8	128148	2024-07-15 14:14:13.970606
36	10	\N	3	128148	2024-07-15 14:14:13.971402
37	10	\N	12	128530	2024-07-15 14:14:13.971891
\.


--
-- Data for Name: post; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.post (id, user_id, media_urls, description, type, created_at) FROM stdin;
4	10	{}	This is a post mentioning @itz_butcher and hashtaging #willy.	video	2024-07-14 21:04:14.940727
5	12	{}	Butcher likes to call people "cunt" 	photo	2024-07-14 21:14:19.660473
8	11	{}	Pitch Perfect 2 is one of the best movies I starred in.	photo	2024-07-15 08:53:54.629582
9	11	{}	Becca: "Anything on radio, basically, right?"	photo	2024-07-15 08:55:21.508695
10	13	{}	Mom! You knew all this is the work of Compound V???	photo	2024-07-15 08:57:35.348818
\.


--
-- Data for Name: reported_message; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.reported_message (id, reporting_user_id, reported_user_id, message_id, reason, reported_at) FROM stdin;
\.


--
-- Data for Name: repost; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.repost (id, reposter_user_id, post_id) FROM stdin;
1	11	4
8	13	5
3	13	9
4	10	9
5	12	4
6	12	8
7	12	9
9	11	5
10	10	8
11	13	8
12	10	5
13	12	10
14	11	10
15	10	10
\.


--
-- Data for Name: saved_post; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.saved_post (id, saver_user_id, post_id) FROM stdin;
3	12	9
4	13	4
2	12	4
6	13	9
5	10	9
7	13	8
8	10	8
9	11	5
10	12	8
11	13	5
12	10	5
13	12	10
14	11	10
15	10	10
\.


--
-- Data for Name: user_conversation; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_conversation (id, user_id, conversation_id, unread_messages_count, notification_mode, deleted, updated_at, partner_user_id) FROM stdin;
2	12	1	1	enabled	f	2024-07-15 21:36:38.113	10
1	10	1	0	enabled	f	2024-07-15 21:18:32.031916	12
\.


--
-- Name: blocked_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.blocked_user_id_seq', 1, false);


--
-- Name: comment_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.comment_id_seq', 45, true);


--
-- Name: conversation_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.conversation_id_seq', 1, true);


--
-- Name: follow_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.follow_id_seq', 83, true);


--
-- Name: i9l_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.i9l_user_id_seq', 13, true);


--
-- Name: message_deletion_log_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.message_deletion_log_id_seq', 2, true);


--
-- Name: message_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.message_id_seq', 7, true);


--
-- Name: message_reaction_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.message_reaction_id_seq', 3, true);


--
-- Name: notification_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.notification_id_seq', 290, true);


--
-- Name: pc_hashtag_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pc_hashtag_id_seq', 2, true);


--
-- Name: pc_mention_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pc_mention_id_seq', 106, true);


--
-- Name: pc_reaction_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pc_reaction_id_seq', 37, true);


--
-- Name: post_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.post_id_seq', 10, true);


--
-- Name: reported_message_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.reported_message_id_seq', 1, false);


--
-- Name: repost_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.repost_id_seq', 15, true);


--
-- Name: saved_post_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.saved_post_id_seq', 15, true);


--
-- Name: user_conversation_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_conversation_id_seq', 2, true);


--
-- Name: blocked_user BlockedUser_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.blocked_user
    ADD CONSTRAINT "BlockedUser_pkey" PRIMARY KEY (id);


--
-- Name: comment_ Comment_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_
    ADD CONSTRAINT "Comment_pkey" PRIMARY KEY (id);


--
-- Name: conversation Conversation_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.conversation
    ADD CONSTRAINT "Conversation_pkey" PRIMARY KEY (id);


--
-- Name: follow FollowAction_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT "FollowAction_pkey" PRIMARY KEY (id);


--
-- Name: pc_hashtag Hashtag_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_hashtag
    ADD CONSTRAINT "Hashtag_pkey" PRIMARY KEY (id);


--
-- Name: pc_mention Mention_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_mention
    ADD CONSTRAINT "Mention_pkey" PRIMARY KEY (id);


--
-- Name: message_reaction MessageReaction_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_reaction
    ADD CONSTRAINT "MessageReaction_pkey" PRIMARY KEY (id);


--
-- Name: message_ Message_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_
    ADD CONSTRAINT "Message_pkey" PRIMARY KEY (id);


--
-- Name: notification Notification_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT "Notification_pkey" PRIMARY KEY (id);


--
-- Name: pc_hashtag PostCommentHashtag_hashtag_name_comment_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_hashtag
    ADD CONSTRAINT "PostCommentHashtag_hashtag_name_comment_id_key" UNIQUE (hashtag_name, comment_id);


--
-- Name: pc_hashtag PostCommentHashtag_hashtag_name_post_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_hashtag
    ADD CONSTRAINT "PostCommentHashtag_hashtag_name_post_id_key" UNIQUE (hashtag_name, post_id);


--
-- Name: pc_mention PostCommentMention_user_id_comment_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_mention
    ADD CONSTRAINT "PostCommentMention_user_id_comment_id_key" UNIQUE (user_id, comment_id);


--
-- Name: pc_mention PostCommentMention_user_id_post_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_mention
    ADD CONSTRAINT "PostCommentMention_user_id_post_id_key" UNIQUE (user_id, post_id);


--
-- Name: pc_reaction Reaction_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_reaction
    ADD CONSTRAINT "Reaction_pkey" PRIMARY KEY (id);


--
-- Name: reported_message ReportedMesssage_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reported_message
    ADD CONSTRAINT "ReportedMesssage_pkey" PRIMARY KEY (id);


--
-- Name: repost Repost_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repost
    ADD CONSTRAINT "Repost_pkey" PRIMARY KEY (id);


--
-- Name: saved_post SavedPost_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saved_post
    ADD CONSTRAINT "SavedPost_pkey" PRIMARY KEY (id);


--
-- Name: user_conversation UserConversation_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_conversation
    ADD CONSTRAINT "UserConversation_pkey" PRIMARY KEY (id);


--
-- Name: i9l_user User_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.i9l_user
    ADD CONSTRAINT "User_pkey" PRIMARY KEY (id);


--
-- Name: blocked_user blocking_is_once_per_user; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.blocked_user
    ADD CONSTRAINT blocking_is_once_per_user UNIQUE (blocking_user_id, blocked_user_id);


--
-- Name: follow follow_is_once; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT follow_is_once UNIQUE (follower_user_id, followee_user_id);


--
-- Name: message_deletion_log message_deletion_log_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_deletion_log
    ADD CONSTRAINT message_deletion_log_pkey PRIMARY KEY (id);


--
-- Name: pc_reaction one_comment_reaction_per_user; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_reaction
    ADD CONSTRAINT one_comment_reaction_per_user UNIQUE (reactor_user_id, target_comment_id);


--
-- Name: pc_reaction one_post_reaction_per_user; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_reaction
    ADD CONSTRAINT one_post_reaction_per_user UNIQUE (reactor_user_id, target_post_id);


--
-- Name: saved_post one_post_save_per_user; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saved_post
    ADD CONSTRAINT one_post_save_per_user UNIQUE (saver_user_id, post_id);


--
-- Name: post post_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT post_pkey PRIMARY KEY (id);


--
-- Name: repost repost_once; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repost
    ADD CONSTRAINT repost_once UNIQUE (reposter_user_id, post_id);


--
-- Name: saved_post save_once; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saved_post
    ADD CONSTRAINT save_once UNIQUE (saver_user_id, post_id);


--
-- Name: ongoing_registration session_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ongoing_registration
    ADD CONSTRAINT session_pkey PRIMARY KEY (sid);


--
-- Name: i9l_user unique_email; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.i9l_user
    ADD CONSTRAINT unique_email UNIQUE (email);


--
-- Name: i9l_user unique_username; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.i9l_user
    ADD CONSTRAINT unique_username UNIQUE (username);


--
-- Name: user_conversation userX_to_conversationX_once; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_conversation
    ADD CONSTRAINT "userX_to_conversationX_once" UNIQUE (user_id, conversation_id);


--
-- Name: message_reaction user_reacts_once; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_reaction
    ADD CONSTRAINT user_reacts_once UNIQUE (message_id, reactor_user_id);


--
-- Name: IDX_session_expire; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_session_expire" ON public.ongoing_registration USING btree (expire);


--
-- Name: blocked_user blocked_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.blocked_user
    ADD CONSTRAINT blocked_user FOREIGN KEY (blocked_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: blocked_user blocking_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.blocked_user
    ADD CONSTRAINT blocking_user FOREIGN KEY (blocking_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: comment_ comment_by; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_
    ADD CONSTRAINT comment_by FOREIGN KEY (commenter_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: comment_ comment_commented_on; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_
    ADD CONSTRAINT comment_commented_on FOREIGN KEY (target_comment_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: notification comment_created; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT comment_created FOREIGN KEY (comment_created_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: pc_mention comment_mentioned_in; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_mention
    ADD CONSTRAINT comment_mentioned_in FOREIGN KEY (comment_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: pc_reaction comment_reacted_to; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_reaction
    ADD CONSTRAINT comment_reacted_to FOREIGN KEY (target_comment_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: conversation conversation_initiator_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.conversation
    ADD CONSTRAINT conversation_initiator_user_id_fkey FOREIGN KEY (initiator_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: conversation conversation_with_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.conversation
    ADD CONSTRAINT conversation_with_user_id_fkey FOREIGN KEY (with_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: user_conversation convo_participant; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_conversation
    ADD CONSTRAINT convo_participant FOREIGN KEY (user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: follow followed_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT followed_user FOREIGN KEY (followee_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: follow follower_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT follower_user FOREIGN KEY (follower_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: pc_hashtag hashtaged_comment; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_hashtag
    ADD CONSTRAINT hashtaged_comment FOREIGN KEY (comment_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: pc_hashtag hashtaged_post; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_hashtag
    ADD CONSTRAINT hashtaged_post FOREIGN KEY (post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: message_deletion_log message_deletion_log_deleter_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_deletion_log
    ADD CONSTRAINT message_deletion_log_deleter_user_id_fkey FOREIGN KEY (deleter_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: message_deletion_log message_deletion_log_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_deletion_log
    ADD CONSTRAINT message_deletion_log_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.message_(id) ON DELETE CASCADE;


--
-- Name: message_reaction message_reacted_to; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_reaction
    ADD CONSTRAINT message_reacted_to FOREIGN KEY (message_id) REFERENCES public.message_(id) ON DELETE CASCADE;


--
-- Name: reported_message message_reported; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reported_message
    ADD CONSTRAINT message_reported FOREIGN KEY (message_id) REFERENCES public.message_(id) ON DELETE CASCADE;


--
-- Name: message_ message_sender; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_
    ADD CONSTRAINT message_sender FOREIGN KEY (sender_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: notification notification_receiver; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT notification_receiver FOREIGN KEY (receiver_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: notification notification_sender; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT notification_sender FOREIGN KEY (sender_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: message_ owner_conversation; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_
    ADD CONSTRAINT owner_conversation FOREIGN KEY (conversation_id) REFERENCES public.conversation(id) ON DELETE CASCADE;


--
-- Name: user_conversation owner_conversation; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_conversation
    ADD CONSTRAINT owner_conversation FOREIGN KEY (conversation_id) REFERENCES public.conversation(id) ON DELETE CASCADE;


--
-- Name: comment_ post_commented_on; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_
    ADD CONSTRAINT post_commented_on FOREIGN KEY (target_post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: pc_mention post_mentioned_in; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_mention
    ADD CONSTRAINT post_mentioned_in FOREIGN KEY (post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: pc_reaction post_reacted_to; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_reaction
    ADD CONSTRAINT post_reacted_to FOREIGN KEY (target_post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: saved_post post_saver; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saved_post
    ADD CONSTRAINT post_saver FOREIGN KEY (saver_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: post posted_by; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT posted_by FOREIGN KEY (user_id) REFERENCES public.i9l_user(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: pc_reaction reaction_by; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_reaction
    ADD CONSTRAINT reaction_by FOREIGN KEY (reactor_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: message_reaction reactor_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_reaction
    ADD CONSTRAINT reactor_user FOREIGN KEY (reactor_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: message_ replied_message; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.message_
    ADD CONSTRAINT replied_message FOREIGN KEY (reply_to_id) REFERENCES public.message_(id) ON DELETE CASCADE;


--
-- Name: reported_message reported_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reported_message
    ADD CONSTRAINT reported_user FOREIGN KEY (reported_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: reported_message reporting_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reported_message
    ADD CONSTRAINT reporting_user FOREIGN KEY (reporting_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: repost reposted_post; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repost
    ADD CONSTRAINT reposted_post FOREIGN KEY (post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: repost reposter; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repost
    ADD CONSTRAINT reposter FOREIGN KEY (reposter_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: saved_post saved_post; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saved_post
    ADD CONSTRAINT saved_post FOREIGN KEY (post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: notification through_comment; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT through_comment FOREIGN KEY (via_comment_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: notification through_post; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT through_post FOREIGN KEY (via_post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: user_conversation user_conversation_partner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_conversation
    ADD CONSTRAINT user_conversation_partner_user_id_fkey FOREIGN KEY (partner_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: pc_mention user_mentioned; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pc_mention
    ADD CONSTRAINT user_mentioned FOREIGN KEY (user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

