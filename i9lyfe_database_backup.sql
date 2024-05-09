--
-- PostgreSQL database dump
--

-- Dumped from database version 16.1
-- Dumped by pg_dump version 16.1

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
  UPDATE user_conversation SET unread_messages_count = unread_messages_count + 1, updated_at = delivery_time
  WHERE user_id = client_user_id AND conversation_id = in_conversation_id
  RETURNING partner_user_id INTO convo_partner_user_id;
  
  -- convo_partner_user_id is a "guard" condition -- the message you ack must indeed belong to your conversation partner
  UPDATE message_ SET delivery_status = 'delivered'
  WHERE id = in_message_id AND conversation_id = in_conversation_id AND sender_user_id = convo_partner_user_id;
  
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
  UPDATE user_conversation SET unread_messages_count = CASE WHEN unread_messages_count > 0 THEN unread_messages_count - 1 ELSE 0 END
  WHERE user_id = client_user_id AND conversation_id = in_conversation_id
  RETURNING partner_user_id INTO convo_partner_user_id;
  
  -- convo_partner_user_id is a "guard" condition -- the message you ack must indeed belong to your conversation partner
  UPDATE message_ SET delivery_status = 'read'
  WHERE id = in_message_id AND conversation_id = in_conversation_id AND sender_user_id = convo_partner_user_id;
  
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
	
	-- create mentions
    INSERT INTO pc_mention (comment_id, user_id)
	VALUES (ret_comment_id, ment_user_id);
	
	-- skip mention notification for client user
	CONTINUE WHEN ment_user_id = client_user_id;
	
	-- create mention notifications
	INSERT INTO notification (type, sender_user_id, receiver_user_id, comment_id)
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
  INSERT INTO notification (type, sender_user_id, receiver_user_id, comment_id, comment_created_id)
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
	
	-- create mentions
    INSERT INTO pc_mention (comment_id, user_id)
	VALUES (ret_comment_id, ment_user_id);
	
	-- skip mention notification for client user
	CONTINUE WHEN ment_user_id = client_user_id;
	
	-- create mention notifications
	INSERT INTO notification (type, sender_user_id, receiver_user_id, comment_id)
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
  INSERT INTO notification (type, sender_user_id, receiver_user_id, post_id, comment_created_id)
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
	
	-- create mentions
    INSERT INTO pc_mention (post_id, user_id)
	VALUES (ret_post_id, ment_user_id);
	
	-- skip mention notification for client user
	CONTINUE WHEN ment_user_id = client_user_id;
	
	-- create mention notifications
	INSERT INTO notification (type, sender_user_id, receiver_user_id, post_id)
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
  
  INSERT INTO notification (type, sender_user_id, receiver_user_id, comment_id, reaction_code_point)
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
  
  INSERT INTO notification (type, sender_user_id, receiver_user_id, post_id, reaction_code_point)
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

CREATE FUNCTION public.create_user(OUT new_user json, in_email character varying, in_username character varying, in_password character varying, in_name character varying, in_birthday timestamp without time zone, in_bio character varying) RETURNS json
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO i9l_user(email, username, password, name, birthday, bio)
  VALUES(in_email, in_username, in_password, in_name, in_birthday, in_bio) 
  RETURNING json_build_object(
	  'id', id, 
	  'email', email, 
	  'username', username, 
	  'name', name, 
	  'profile_pic_url', profile_pic_url,
	  'connection_status', connection_status
  ) INTO new_user;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_user(OUT new_user json, in_email character varying, in_username character varying, in_password character varying, in_name character varying, in_birthday timestamp without time zone, in_bio character varying) OWNER TO postgres;

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
    IF col_name_val[1] NOT IN ('username', 'password', 'email', 'name', 'profile_picture', 'birthday', 'bio') THEN
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

CREATE FUNCTION public.get_user(OUT res_user json, unique_identifier character varying) RETURNS json
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT json_build_object(
	  'id', id, 
	  'email', email, 
	  'username', username, 
	  'name', name, 
	  'profile_pic_url', profile_pic_url,
	  'connection_status', connection_status,
	  'password', password
  ) INTO res_user FROM i9l_user
  WHERE unique_identifier = ANY(ARRAY[id::varchar, email, username]);
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_user(OUT res_user json, unique_identifier character varying) OWNER TO postgres;

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
      ON client_follows.followee_user_id = followee_user.id AND client_follows.follower_user_id = client_user_id
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

CREATE FUNCTION public.get_user_notifications(OUT user_notifications json, client_user_id integer, in_from timestamp without time zone, in_limit integer, in_offset integer) RETURNS json
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT json_agg(notif) INTO user_notifications FROM (SELECT json_strip_nulls(json_build_object(
	  'type', n.type,
	  'is_read', n.is_read,
	  'sender', json_build_object(
		  'id', sender.id,
		  'username', sender.username,
		  'profile_pic_url', sender.profile_pic_url
	  ),
	  'post_id', n.post_id,
	  'comment_id', n.comment_id,
	  'comment_created_id', n.comment_created_id,
	  'reaction_code_point', n.reaction_code_point,
	  'created_at', n.created_at
  )) AS notif FROM notification n
  INNER JOIN i9l_user sender ON sender.id = n.sender_user_id
  WHERE n.receiver_user_id = client_user_id AND n.created_at >= in_from
  ORDER BY n.created_at DESC
  LIMIT in_limit OFFSET in_offset);
  
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_user_notifications(OUT user_notifications json, client_user_id integer, in_from timestamp without time zone, in_limit integer, in_offset integer) OWNER TO postgres;

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

CREATE FUNCTION public.get_user_profile(OUT profile_data json, in_username character varying, client_user_id integer) RETURNS json
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT json_build_object(
	  'user_id', i9l_user.id,
      'name', name,
      'username', username,
      'bio', bio,
      'profile_pic_url', profile_pic_url, 
      'followers_count', COUNT(followee.id), 
      'following_count', COUNT(follower.id),
      'client_follows', CASE
        WHEN client_follows.id IS NULL THEN false
        ELSE true
      END
  ) INTO profile_data
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


ALTER FUNCTION public.get_user_profile(OUT profile_data json, in_username character varying, client_user_id integer) OWNER TO postgres;

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
    post_id integer,
    comment_id integer,
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
    ADD CONSTRAINT through_comment FOREIGN KEY (comment_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: notification through_post; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT through_post FOREIGN KEY (post_id) REFERENCES public.post(id) ON DELETE CASCADE;


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

