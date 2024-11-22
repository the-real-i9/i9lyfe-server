--
-- PostgreSQL database dump
--

-- Dumped from database version 17.0 (Ubuntu 17.0-1.pgdg24.04+1)
-- Dumped by pg_dump version 17.0 (Ubuntu 17.0-1.pgdg24.04+1)

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
-- Name: i9l_user_profile_t; Type: TYPE; Schema: public; Owner: i9
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


ALTER TYPE public.i9l_user_profile_t OWNER TO i9;

--
-- Name: i9l_user_t; Type: TYPE; Schema: public; Owner: i9
--

CREATE TYPE public.i9l_user_t AS (
	id integer,
	email character varying,
	username character varying,
	name character varying,
	profile_pic_url character varying,
	connection_status text
);


ALTER TYPE public.i9l_user_t OWNER TO i9;

--
-- Name: ui_comment_struct; Type: TYPE; Schema: public; Owner: i9
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


ALTER TYPE public.ui_comment_struct OWNER TO i9;

--
-- Name: ui_post_struct; Type: TYPE; Schema: public; Owner: i9
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


ALTER TYPE public.ui_post_struct OWNER TO i9;

--
-- Name: ack_msg_delivered(integer, integer, integer, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.ack_msg_delivered(client_user_id integer, in_chat_id integer, in_message_id integer, delivery_time timestamp without time zone) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
  chat_partner_user_id int;
BEGIN
IF (SELECT delivery_status FROM message_ WHERE id = in_message_id) <> 'delivered' THEN
  UPDATE user_chat SET unread_messages_count = unread_messages_count + 1, updated_at = delivery_time
  WHERE user_id = client_user_id AND chat_id = in_chat_id
  RETURNING partner_user_id INTO chat_partner_user_id;
  
  -- chat_partner_user_id is a "guard" condition asserting that the message you ack must indeed belong to your chat partner
  UPDATE message_ SET delivery_status = 'delivered'
  WHERE id = in_message_id AND chat_id = in_chat_id AND sender_user_id = chat_partner_user_id;
END IF;
  
  RETURN true;
END;
$$;


ALTER FUNCTION public.ack_msg_delivered(client_user_id integer, in_chat_id integer, in_message_id integer, delivery_time timestamp without time zone) OWNER TO i9;

--
-- Name: ack_msg_read(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.ack_msg_read(client_user_id integer, in_chat_id integer, in_message_id integer) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
  chat_partner_user_id int;
BEGIN
IF (SELECT delivery_status FROM message_ WHERE id = in_message_id) <> 'read' THEN
  UPDATE user_chat SET unread_messages_count = CASE WHEN unread_messages_count > 0 THEN unread_messages_count - 1 ELSE 0 END
  WHERE user_id = client_user_id AND chat_id = in_chat_id
  RETURNING partner_user_id INTO chat_partner_user_id;
  
  -- chat_partner_user_id is a "guard" condition asserting that the message you ack must indeed belong to your chat partner
  UPDATE message_ SET delivery_status = 'read'
  WHERE id = in_message_id AND chat_id = in_chat_id AND sender_user_id = chat_partner_user_id;
END IF;
 
  RETURN true;
END;
$$;


ALTER FUNCTION public.ack_msg_read(client_user_id integer, in_chat_id integer, in_message_id integer) OWNER TO i9;

--
-- Name: change_password(character varying, character varying); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.change_password(in_email character varying, in_new_password character varying) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  UPDATE i9l_user SET password = in_new_password WHERE email = in_email;
  
  RETURN true;
END;
$$;


ALTER FUNCTION public.change_password(in_email character varying, in_new_password character varying) OWNER TO i9;

--
-- Name: create_comment_on_comment(integer, integer, integer, text, text, character varying[], character varying[]); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.create_comment_on_comment(OUT new_comment_data json, OUT comment_notif json, OUT mention_notifs json[], OUT latest_comments_count integer, in_comment_id integer, comment_owner_user_id integer, client_user_id integer, in_comment_text text, in_attachment_url text, mentions character varying[], hashtags character varying[]) RETURNS record
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
  INSERT INTO comment_ (comment_id, commenter_user_id, comment_text, attachment_url)
  VALUES (in_comment_id, client_user_id, in_comment_text, in_attachment_url)
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
  VALUES ('comment_on_comment', client_user_id, comment_owner_user_id, in_comment_id, ret_comment_id);
  
  
  new_comment_data := json_build_object(
    'owner_user', client_data,
	'comment_id', ret_comment_id,
	'attachment_url', in_attachment_url,
	'comment_text', in_comment_text,
	'reactions_count', 0,
    'comments_count', 0,
	'client_reaction', ''
  );
  mention_notifs := mention_notifs_acc;
  comment_notif := json_build_object(
	  'receiver_user_id', comment_owner_user_id,
	  'type', 'comment_on_comment',
	  'sender', client_data,
	  'comment_id', in_comment_id,
	  'comment_created_id', ret_comment_id
  );
  
  SELECT COUNT(1) + 1 INTO latest_comments_count FROM comment_ WHERE comment_id = in_comment_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_comment_on_comment(OUT new_comment_data json, OUT comment_notif json, OUT mention_notifs json[], OUT latest_comments_count integer, in_comment_id integer, comment_owner_user_id integer, client_user_id integer, in_comment_text text, in_attachment_url text, mentions character varying[], hashtags character varying[]) OWNER TO i9;

--
-- Name: create_comment_on_post(integer, integer, integer, text, text, character varying[], character varying[]); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.create_comment_on_post(OUT new_comment_data json, OUT comment_notif json, OUT mention_notifs json[], OUT latest_comments_count integer, in_post_id integer, post_owner_user_id integer, client_user_id integer, in_comment_text text, in_attachment_url text, mentions character varying[], hashtags character varying[]) RETURNS record
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
  INSERT INTO comment_ (post_id, commenter_user_id, comment_text, attachment_url)
  VALUES (in_post_id, client_user_id, in_comment_text, in_attachment_url)
  RETURNING id INTO ret_comment_id;
  
  -- populate client data
  SELECT json_build_object(
	  'user_id', id,
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
  VALUES ('comment_on_post', client_user_id, post_owner_user_id, in_post_id, ret_comment_id);
  
  
  new_comment_data := json_build_object(
    'owner_user', client_data,
	'comment_id', ret_comment_id,
	'attachment_url', in_attachment_url,
	'comment_text', in_comment_text,
	'reactions_count', 0,
    'comments_count', 0,
	'client_reaction', ''
  );
  mention_notifs := mention_notifs_acc;
  comment_notif := json_build_object(
	  'receiver_user_id', post_owner_user_id,
	  'type', 'comment_on_post',
	  'sender', client_data,
	  'post_id', in_post_id,
	  'comment_created_id', ret_comment_id
  );
  
  SELECT COUNT(1) + 1 INTO latest_comments_count FROM comment_ WHERE post_id = in_post_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_comment_on_post(OUT new_comment_data json, OUT comment_notif json, OUT mention_notifs json[], OUT latest_comments_count integer, in_post_id integer, post_owner_user_id integer, client_user_id integer, in_comment_text text, in_attachment_url text, mentions character varying[], hashtags character varying[]) OWNER TO i9;

--
-- Name: create_chat(integer, integer, json); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.create_chat(OUT client_res json, OUT partner_res json, in_initiator_user_id integer, in_with_user_id integer, init_message json) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  ret_chat_id int;
  ret_message_id int;
  
  client_data json;
BEGIN
  INSERT INTO chat(initiator_user_id, with_user_id)
  VALUES (in_initiator_user_id, in_with_user_id)
  RETURNING id INTO ret_chat_id;
  
  INSERT INTO user_chat(chat_id, user_id, partner_user_id)
  VALUES (ret_chat_id, in_initiator_user_id, in_with_user_id);
  
  INSERT INTO user_chat(chat_id, user_id, partner_user_id)
  VALUES (ret_chat_id, in_with_user_id, in_initiator_user_id);
  
  INSERT INTO message_(sender_user_id, chat_id, msg_content)
  VALUES (in_initiator_user_id, ret_chat_id, init_message)
  RETURNING id INTO ret_message_id;
  
  SELECT json_build_object('username', username, 'profile_pic_url', profile_pic_url) INTO client_data
  FROM i9l_user WHERE id = in_initiator_user_id;
  
  client_res := json_build_object('chat_id', ret_chat_id, 'init_message_id', ret_message_id);
  
  partner_res := json_build_object(
	  'chat', json_build_object(
		  'id', ret_chat_id,
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


ALTER FUNCTION public.create_chat(OUT client_res json, OUT partner_res json, in_initiator_user_id integer, in_with_user_id integer, init_message json) OWNER TO i9;

--
-- Name: create_message(integer, integer, json); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.create_message(OUT client_res json, OUT partner_res json, in_chat_id integer, client_user_id integer, in_msg_content json) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  ret_msg_id int;
  
  client_data json;
BEGIN
  INSERT INTO message_ (sender_user_id, chat_id, msg_content) 
  VALUES (client_user_id, in_chat_id, in_msg_content)
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
	  'chat_id', in_chat_id,
	  'new_msg_id', ret_msg_id,
	  'sender', client_data,
	  'msg_content', in_msg_content
  );
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_message(OUT client_res json, OUT partner_res json, in_chat_id integer, client_user_id integer, in_msg_content json) OWNER TO i9;

--
-- Name: create_post(integer, text[], text, text, character varying[], character varying[]); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.create_post(OUT new_post_data json, OUT mention_notifs json[], client_user_id integer, in_media_urls text[], in_type text, in_description text, mentions character varying[], hashtags character varying[]) RETURNS record
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
	  'user_id', id,
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
  
  
  
  new_post_data := json_build_object(
    'owner_user', client_data,
	'post_id', ret_post_id,
	'type', in_type,
	'media_urls', in_media_urls,
	'description', in_description,
	'reactions_count', 0,
    'comments_count', 0,
    'reposts_count', 0,
    'saves_count', 0,
	'client_reaction', '',
	'client_reposted', false,
	'client_saved', false
  );
  mention_notifs := mention_notifs_acc;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_post(OUT new_post_data json, OUT mention_notifs json[], client_user_id integer, in_media_urls text[], in_type text, in_description text, mentions character varying[], hashtags character varying[]) OWNER TO i9;

--
-- Name: create_reaction_to_comment(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.create_reaction_to_comment(OUT reaction_notif json, OUT latest_reactions_count integer, client_user_id integer, in_comment_id integer, comment_owner_user_id integer, in_reaction_code_point integer) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  client_data json;
BEGIN
  INSERT INTO pc_reaction (reactor_user_id, comment_id, reaction_code_point)
  VALUES (client_user_id, in_comment_id, in_reaction_code_point);
  
  -- populate client data
  SELECT json_build_object(
	  'id', id,
	  'username', username,
	  'profile_pic_url', profile_pic_url
  ) INTO client_data FROM i9l_user WHERE id = client_user_id;
  
  INSERT INTO notification (type, sender_user_id, receiver_user_id, via_comment_id, reaction_code_point)
  VALUES ('reaction_to_comment', client_user_id, comment_owner_user_id, in_comment_id, in_reaction_code_point);
  
  reaction_notif := json_build_object(
	  'receiver_user_id', comment_owner_user_id,
	  'type', 'reaction_to_comment',
	  'reaction_code_point', in_reaction_code_point,
	  'comment_id', in_comment_id,
	  'sender', client_data
	  
  );
  
  SELECT COUNT(1) + 1 INTO latest_reactions_count FROM pc_reaction WHERE comment_id = in_comment_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_reaction_to_comment(OUT reaction_notif json, OUT latest_reactions_count integer, client_user_id integer, in_comment_id integer, comment_owner_user_id integer, in_reaction_code_point integer) OWNER TO i9;

--
-- Name: create_reaction_to_post(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.create_reaction_to_post(OUT reaction_notif json, OUT latest_reactions_count integer, client_user_id integer, in_post_id integer, post_owner_user_id integer, in_reaction_code_point integer) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  client_data json;
BEGIN
  INSERT INTO pc_reaction (reactor_user_id, post_id, reaction_code_point)
  VALUES (client_user_id, in_post_id, in_reaction_code_point);
  
  -- populate client data
  SELECT json_build_object(
	  'id', id,
	  'username', username,
	  'profile_pic_url', profile_pic_url
  ) INTO client_data FROM i9l_user WHERE id = client_user_id;
  
  INSERT INTO notification (type, sender_user_id, receiver_user_id, via_post_id, reaction_code_point)
  VALUES ('reaction_to_post', client_user_id, post_owner_user_id, in_post_id, in_reaction_code_point);
  
  reaction_notif := json_build_object(
	  'receiver_user_id', post_owner_user_id,
	  'type', 'reaction_to_post',
	  'post_id', in_post_id,
	  'reaction_code_point', in_reaction_code_point,
	  'sender', client_data
  );
  
  SELECT COUNT(1) + 1 INTO latest_reactions_count FROM pc_reaction WHERE post_id = in_post_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.create_reaction_to_post(OUT reaction_notif json, OUT latest_reactions_count integer, client_user_id integer, in_post_id integer, post_owner_user_id integer, in_reaction_code_point integer) OWNER TO i9;

--
-- Name: create_user(character varying, character varying, character varying, character varying, timestamp without time zone, character varying); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.create_user(in_email character varying, in_username character varying, in_password character varying, in_name character varying, in_birthday timestamp without time zone, in_bio character varying) OWNER TO i9;

--
-- Name: edit_user(integer, character varying[]); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.edit_user(client_user_id integer, col_updates character varying[]) OWNER TO i9;

--
-- Name: fetch_home_feed_posts(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.fetch_home_feed_posts(client_user_id integer, in_limit integer, in_offset integer) RETURNS SETOF public.ui_post_struct
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- This stored function aggregates posts based on the "content recommendation algorithm"
  
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
    WHERE owner_user_id = client_user_id OR recommend_post(client_user_id, post_id)
    ORDER BY created_at DESC
    LIMIT in_limit OFFSET in_offset;
	  
END;
$$;


ALTER FUNCTION public.fetch_home_feed_posts(client_user_id integer, in_limit integer, in_offset integer) OWNER TO i9;

--
-- Name: follow_user(integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.follow_user(OUT follow_notif json, client_user_id integer, to_follow_user_id integer) OWNER TO i9;

--
-- Name: get_comment(integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_comment(in_comment_id integer, client_user_id integer) OWNER TO i9;

--
-- Name: get_comments_on_comment(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.get_comments_on_comment(in_comment_id integer, client_user_id integer, in_limit integer, in_offset integer) RETURNS SETOF public.ui_comment_struct
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
    WHERE comment_id = in_comment_id
    ORDER BY created_at DESC
    LIMIT in_limit OFFSET in_offset;
	  
END;
$$;


ALTER FUNCTION public.get_comments_on_comment(in_comment_id integer, client_user_id integer, in_limit integer, in_offset integer) OWNER TO i9;

--
-- Name: get_comments_on_post(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.get_comments_on_post(in_post_id integer, client_user_id integer, in_limit integer, in_offset integer) RETURNS SETOF public.ui_comment_struct
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
    WHERE post_id = in_post_id
    ORDER BY created_at DESC
    LIMIT in_limit OFFSET in_offset;
	  
END;
$$;


ALTER FUNCTION public.get_comments_on_post(in_post_id integer, client_user_id integer, in_limit integer, in_offset integer) OWNER TO i9;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: i9l_user; Type: TABLE; Schema: public; Owner: i9
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


ALTER TABLE public.i9l_user OWNER TO i9;

--
-- Name: message_; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.message_ (
    id integer NOT NULL,
    sender_user_id integer NOT NULL,
    chat_id integer NOT NULL,
    msg_content jsonb,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    delivery_status text DEFAULT 'sent'::text NOT NULL,
    reply_to_id integer,
    CONSTRAINT "Message_delivery_status_check" CHECK ((delivery_status = ANY (ARRAY['sent'::text, 'delivered'::text, 'read'::text])))
);


ALTER TABLE public.message_ OWNER TO i9;

--
-- Name: message_reaction; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.message_reaction (
    id integer NOT NULL,
    message_id integer NOT NULL,
    reactor_user_id integer NOT NULL,
    reaction_code_point integer NOT NULL
);


ALTER TABLE public.message_reaction OWNER TO i9;

--
-- Name: ChatHistoryView; Type: VIEW; Schema: public; Owner: i9
--

CREATE VIEW public."ChatHistoryView" AS
 SELECT msg.id AS msg_id,
    json_build_object('id', sender.id, 'username', sender.username, 'profile_pic_url', sender.profile_pic_url) AS sender,
    msg.msg_content,
    msg.delivery_status,
    ( SELECT array_agg(message_reaction.reaction_code_point) AS array_agg
           FROM public.message_reaction
          WHERE (message_reaction.message_id = msg.id)) AS reactions,
    msg.created_at,
    msg.chat_id
   FROM (public.message_ msg
     JOIN public.i9l_user sender ON ((sender.id = msg.sender_user_id)))
  ORDER BY msg.created_at DESC;


ALTER VIEW public."ChatHistoryView" OWNER TO i9;

--
-- Name: get_chat_history(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.get_chat_history(in_chat_id integer, in_limit integer, in_offset integer) RETURNS SETOF public."ChatHistoryView"
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT * FROM (
	  SELECT * FROM "ChatHistoryView"
      WHERE chat_id = in_chat_id
      LIMIT in_limit OFFSET in_offset
  ) ORDER BY created_at ASC;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_chat_history(in_chat_id integer, in_limit integer, in_offset integer) OWNER TO i9;

--
-- Name: get_explore_posts(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_explore_posts(in_limit integer, in_offset integer, client_user_id integer) OWNER TO i9;

--
-- Name: get_hashtag_posts(character varying, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_hashtag_posts(in_hashtag_name character varying, in_limit integer, in_offset integer, client_user_id integer) OWNER TO i9;

--
-- Name: get_mentioned_posts(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_mentioned_posts(in_limit integer, in_offset integer, client_user_id integer) OWNER TO i9;

--
-- Name: get_post(integer, integer, boolean); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.get_post(in_post_id integer, client_user_id integer, if_recommended boolean) RETURNS SETOF public.ui_post_struct
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
    WHERE post_id = in_post_id AND (CASE WHEN if_recommended THEN recommend_post(client_user_id, post_id) ELSE true END);
	  
	  
END;
$$;


ALTER FUNCTION public.get_post(in_post_id integer, client_user_id integer, if_recommended boolean) OWNER TO i9;

--
-- Name: get_reacted_posts(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_reacted_posts(in_limit integer, in_offset integer, client_user_id integer) OWNER TO i9;

--
-- Name: get_reactors_to_comment(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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
    WHERE pc_reaction.comment_id = in_comment_id
    ORDER BY pc_reaction.created_at DESC
    LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.get_reactors_to_comment(in_comment_id integer, client_user_id integer, in_limit integer, in_offset integer) OWNER TO i9;

--
-- Name: get_reactors_to_post(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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
    WHERE pc_reaction.post_id = in_post_id
    ORDER BY pc_reaction.created_at DESC
    LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.get_reactors_to_post(in_post_id integer, client_user_id integer, in_limit integer, in_offset integer) OWNER TO i9;

--
-- Name: get_reactors_with_reaction_to_comment(integer, integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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
    WHERE pc_reaction.comment_id = in_comment_id AND pc_reaction.reaction_code_point = in_reaction_code_point
    ORDER BY pc_reaction.created_at DESC
    LIMIT in_limit OFFSET in_offset;
	  
	  
END;
$$;


ALTER FUNCTION public.get_reactors_with_reaction_to_comment(in_comment_id integer, in_reaction_code_point integer, client_user_id integer, in_limit integer, in_offset integer) OWNER TO i9;

--
-- Name: get_reactors_with_reaction_to_post(integer, integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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
    WHERE pc_reaction.post_id = in_post_id AND pc_reaction.reaction_code_point = in_reaction_code_point
    ORDER BY pc_reaction.created_at DESC
    LIMIT in_limit OFFSET in_offset;
	
	END;
$$;


ALTER FUNCTION public.get_reactors_with_reaction_to_post(in_post_id integer, in_reaction_code_point integer, client_user_id integer, in_limit integer, in_offset integer) OWNER TO i9;

--
-- Name: get_saved_posts(integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_saved_posts(in_limit integer, in_offset integer, client_user_id integer) OWNER TO i9;

--
-- Name: get_user(character varying); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_user(unique_identifier character varying) OWNER TO i9;

--
-- Name: get_user_chats(integer); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.get_user_chats(client_user_id integer) RETURNS TABLE(chat_id integer, partner json, unread_messages_count integer, updated_at timestamp without time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT uconv.chat_id,
    json_build_object(
		'id', par.id,
		'username', par.username,
		'profile_pic_url', par.profile_pic_url,
		'connection_status', par.connection_status,
		'last_active', par.last_active
	) AS partner,
    uconv.unread_messages_count,
    uconv.updated_at
  FROM user_chat uconv
  LEFT JOIN i9l_user par ON par.id = uconv.partner_user_id
  WHERE uconv.user_id = client_user_id AND uconv.deleted = false;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_user_chats(client_user_id integer) OWNER TO i9;

--
-- Name: get_user_followers(character varying, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_user_followers(in_username character varying, in_limit integer, in_offset integer, client_user_id integer) OWNER TO i9;

--
-- Name: get_user_following(character varying, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_user_following(in_username character varying, in_limit integer, in_offset integer, client_user_id integer) OWNER TO i9;

--
-- Name: get_user_notifications(integer, timestamp without time zone, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_user_notifications(client_user_id integer, in_from timestamp without time zone, in_limit integer, in_offset integer) OWNER TO i9;

--
-- Name: get_user_password(character varying); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_user_password(OUT pswd character varying, unique_identifier character varying) OWNER TO i9;

--
-- Name: get_user_posts(character varying, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_user_posts(in_username character varying, in_limit integer, in_offset integer, client_user_id integer) OWNER TO i9;

--
-- Name: get_user_profile(character varying, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.get_user_profile(in_username character varying, client_user_id integer) OWNER TO i9;

--
-- Name: get_users_to_chat(text, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.get_users_to_chat(in_search text, in_limit integer, in_offset integer, client_user_id integer) RETURNS TABLE(id integer, username character varying, name character varying, profile_pic_url character varying, connection_status text, chat_id integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT i9l_user.id,
	  i9l_user.username,
	  i9l_user.name,
	  i9l_user.profile_pic_url,
	  i9l_user.connection_status,
	  uconv.chat_id
  FROM i9l_user
  LEFT JOIN user_chat uconv ON uconv.user_id = i9l_user.id
  WHERE (i9l_user.username ILIKE in_search OR i9l_user.name ILIKE in_search) AND i9l_user.id != client_user_id
  LIMIT in_limit OFFSET in_offset;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_users_to_chat(in_search text, in_limit integer, in_offset integer, client_user_id integer) OWNER TO i9;

--
-- Name: recommend_post(integer, integer); Type: FUNCTION; Schema: public; Owner: i9
--

CREATE FUNCTION public.recommend_post(client_user_id integer, post_id integer) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
  post_owner_user_id int;
BEGIN
  -- This is the implementation of the "post recommendation algorithm", 
  -- it promises to be advanced and sophisticated in the near future.
  -- For now, it just checks the following conditions in decreasing order of priority
     -- if the client follows the post owner, else
	 -- if the client follows a user who follows the post owner

  SELECT user_id INTO post_owner_user_id FROM post WHERE id = post_id;

  IF (SELECT EXISTS (SELECT 1 FROM follow 
	WHERE follower_user_id = client_user_id AND followee_user_id = post_owner_user_id)) THEN
    -- does client follow post owner?
	-- rationale: the client is interested in the owner's content
    RETURN true;
  ELSIF (SELECT array_agg(followee_user_id) FROM follow 
	WHERE follower_user_id = client_user_id) @> (SELECT array_agg(follower_user_id) FROM follow 
	WHERE followee_user_id = post_owner_user_id) THEN
	  -- does client follow a user that follows the post owner?
	  -- rationale: possibility of shared interest between the client and the user he follows
      RETURN true;
  END IF;

  RETURN false;
END;
$$;


ALTER FUNCTION public.recommend_post(client_user_id integer, post_id integer) OWNER TO i9;

--
-- Name: search_filter_posts(text, text, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.search_filter_posts(search_text text, filter_text text, in_limit integer, in_offset integer, client_user_id integer) OWNER TO i9;

--
-- Name: user_exists(character varying); Type: FUNCTION; Schema: public; Owner: i9
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


ALTER FUNCTION public.user_exists(OUT check_res boolean, unique_identifier character varying) OWNER TO i9;

--
-- Name: comment_; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.comment_ (
    id integer NOT NULL,
    comment_text text NOT NULL,
    commenter_user_id integer NOT NULL,
    attachment_url text,
    post_id integer,
    comment_id integer,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT either_comment_on_post_or_reply_to_comment CHECK (((post_id IS NULL) OR (comment_id IS NULL)))
);


ALTER TABLE public.comment_ OWNER TO i9;

--
-- Name: pc_reaction; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.pc_reaction (
    id integer NOT NULL,
    reactor_user_id integer NOT NULL,
    post_id integer,
    comment_id integer,
    reaction_code_point integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT reaction_either_in_post_or_comment CHECK (((post_id IS NULL) OR (comment_id IS NULL)))
);


ALTER TABLE public.pc_reaction OWNER TO i9;

--
-- Name: CommentView; Type: VIEW; Schema: public; Owner: i9
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
    cm.post_id,
    cm.comment_id,
    cm.created_at
   FROM ((((public.comment_ cm
     JOIN public.i9l_user ON ((i9l_user.id = cm.commenter_user_id)))
     LEFT JOIN public.pc_reaction any_reaction ON ((any_reaction.comment_id = cm.id)))
     LEFT JOIN public.comment_ cm_on_cm ON ((cm_on_cm.comment_id = cm.id)))
     LEFT JOIN public.pc_reaction certain_reaction ON ((certain_reaction.comment_id = cm.id)))
  GROUP BY i9l_user.id, i9l_user.username, i9l_user.profile_pic_url, cm.id, cm.comment_text, cm.attachment_url, certain_reaction.reactor_user_id, certain_reaction.reaction_code_point, cm.post_id, cm.comment_id, cm.created_at;


ALTER VIEW public."CommentView" OWNER TO i9;

--
-- Name: post; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.post (
    id integer NOT NULL,
    user_id integer NOT NULL,
    media_urls text[] NOT NULL,
    description text,
    type text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.post OWNER TO i9;

--
-- Name: repost; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.repost (
    id integer NOT NULL,
    reposter_user_id integer NOT NULL,
    post_id integer NOT NULL
);


ALTER TABLE public.repost OWNER TO i9;

--
-- Name: saved_post; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.saved_post (
    id integer NOT NULL,
    saver_user_id integer NOT NULL,
    post_id integer NOT NULL
);


ALTER TABLE public.saved_post OWNER TO i9;

--
-- Name: PostView; Type: VIEW; Schema: public; Owner: i9
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
     LEFT JOIN public.pc_reaction any_reaction ON ((any_reaction.post_id = post.id)))
     LEFT JOIN public.comment_ any_comment ON ((any_comment.post_id = post.id)))
     LEFT JOIN public.repost any_repost ON ((any_repost.post_id = post.id)))
     LEFT JOIN public.saved_post any_saved_post ON ((any_saved_post.post_id = post.id)))
     LEFT JOIN public.pc_reaction certain_reaction ON ((certain_reaction.post_id = post.id)))
     LEFT JOIN public.repost certain_repost ON ((certain_repost.post_id = post.id)))
     LEFT JOIN public.saved_post certain_saved_post ON ((certain_saved_post.post_id = post.id)))
  GROUP BY i9l_user.id, i9l_user.username, i9l_user.profile_pic_url, post.id, post.type, post.media_urls, post.description, certain_reaction.reactor_user_id, certain_reaction.reaction_code_point, certain_repost.reposter_user_id, certain_saved_post.saver_user_id, post.created_at;


ALTER VIEW public."PostView" OWNER TO i9;

--
-- Name: blocked_user; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.blocked_user (
    id integer NOT NULL,
    blocking_user_id integer NOT NULL,
    blocked_user_id integer NOT NULL,
    blocked_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.blocked_user OWNER TO i9;

--
-- Name: blocked_user_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.blocked_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.blocked_user_id_seq OWNER TO i9;

--
-- Name: blocked_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.blocked_user_id_seq OWNED BY public.blocked_user.id;


--
-- Name: comment_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.comment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.comment_id_seq OWNER TO i9;

--
-- Name: comment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.comment_id_seq OWNED BY public.comment_.id;


--
-- Name: chat; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.chat (
    id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    initiator_user_id integer NOT NULL,
    with_user_id integer NOT NULL
);


ALTER TABLE public.chat OWNER TO i9;

--
-- Name: chat_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.chat_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.chat_id_seq OWNER TO i9;

--
-- Name: chat_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.chat_id_seq OWNED BY public.chat.id;


--
-- Name: follow; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.follow (
    id integer NOT NULL,
    follower_user_id integer NOT NULL,
    followee_user_id integer NOT NULL,
    follow_on timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT no_self_follow CHECK ((follower_user_id <> followee_user_id))
);


ALTER TABLE public.follow OWNER TO i9;

--
-- Name: follow_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.follow_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.follow_id_seq OWNER TO i9;

--
-- Name: follow_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.follow_id_seq OWNED BY public.follow.id;


--
-- Name: i9l_user_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.i9l_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.i9l_user_id_seq OWNER TO i9;

--
-- Name: i9l_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.i9l_user_id_seq OWNED BY public.i9l_user.id;


--
-- Name: message_deletion_log; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.message_deletion_log (
    id integer NOT NULL,
    deleter_user_id integer NOT NULL,
    message_id integer NOT NULL,
    deleted_for character varying NOT NULL,
    CONSTRAINT message_deletion_log_deleted_for_check CHECK (((deleted_for)::text = ANY (ARRAY['me'::text, 'everyone'::text])))
);


ALTER TABLE public.message_deletion_log OWNER TO i9;

--
-- Name: message_deletion_log_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.message_deletion_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.message_deletion_log_id_seq OWNER TO i9;

--
-- Name: message_deletion_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.message_deletion_log_id_seq OWNED BY public.message_deletion_log.id;


--
-- Name: message_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.message_id_seq OWNER TO i9;

--
-- Name: message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.message_id_seq OWNED BY public.message_.id;


--
-- Name: message_reaction_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.message_reaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.message_reaction_id_seq OWNER TO i9;

--
-- Name: message_reaction_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.message_reaction_id_seq OWNED BY public.message_reaction.id;


--
-- Name: notification; Type: TABLE; Schema: public; Owner: i9
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


ALTER TABLE public.notification OWNER TO i9;

--
-- Name: notification_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.notification_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.notification_id_seq OWNER TO i9;

--
-- Name: notification_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.notification_id_seq OWNED BY public.notification.id;


--
-- Name: ongoing_registration; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.ongoing_registration (
    sid character varying NOT NULL,
    sess json NOT NULL,
    expire timestamp(6) without time zone NOT NULL
);


ALTER TABLE public.ongoing_registration OWNER TO i9;

--
-- Name: pc_hashtag; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.pc_hashtag (
    id integer NOT NULL,
    post_id integer,
    comment_id integer,
    hashtag_name character varying(255) NOT NULL,
    CONSTRAINT hashtag_either_in_post_or_comment CHECK (((post_id IS NULL) OR (comment_id IS NULL)))
);


ALTER TABLE public.pc_hashtag OWNER TO i9;

--
-- Name: pc_hashtag_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.pc_hashtag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pc_hashtag_id_seq OWNER TO i9;

--
-- Name: pc_hashtag_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.pc_hashtag_id_seq OWNED BY public.pc_hashtag.id;


--
-- Name: pc_mention; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.pc_mention (
    id integer NOT NULL,
    post_id integer,
    comment_id integer,
    user_id integer NOT NULL,
    CONSTRAINT mention_either_in_post_or_comment CHECK (((post_id IS NULL) OR (comment_id IS NULL)))
);


ALTER TABLE public.pc_mention OWNER TO i9;

--
-- Name: pc_mention_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.pc_mention_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pc_mention_id_seq OWNER TO i9;

--
-- Name: pc_mention_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.pc_mention_id_seq OWNED BY public.pc_mention.id;


--
-- Name: pc_reaction_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.pc_reaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pc_reaction_id_seq OWNER TO i9;

--
-- Name: pc_reaction_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.pc_reaction_id_seq OWNED BY public.pc_reaction.id;


--
-- Name: post_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.post_id_seq OWNER TO i9;

--
-- Name: post_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.post_id_seq OWNED BY public.post.id;


--
-- Name: reported_message; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.reported_message (
    id integer NOT NULL,
    reporting_user_id integer NOT NULL,
    reported_user_id integer NOT NULL,
    message_id integer NOT NULL,
    reason text NOT NULL,
    reported_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.reported_message OWNER TO i9;

--
-- Name: reported_message_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.reported_message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.reported_message_id_seq OWNER TO i9;

--
-- Name: reported_message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.reported_message_id_seq OWNED BY public.reported_message.id;


--
-- Name: repost_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.repost_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.repost_id_seq OWNER TO i9;

--
-- Name: repost_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.repost_id_seq OWNED BY public.repost.id;


--
-- Name: saved_post_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.saved_post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.saved_post_id_seq OWNER TO i9;

--
-- Name: saved_post_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.saved_post_id_seq OWNED BY public.saved_post.id;


--
-- Name: user_chat; Type: TABLE; Schema: public; Owner: i9
--

CREATE TABLE public.user_chat (
    id integer NOT NULL,
    user_id integer NOT NULL,
    chat_id integer NOT NULL,
    unread_messages_count integer DEFAULT 0,
    notification_mode text DEFAULT 'enabled'::text NOT NULL,
    deleted boolean DEFAULT false,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    partner_user_id integer NOT NULL,
    CONSTRAINT "UserChat_notification_mode_check" CHECK ((notification_mode = ANY (ARRAY['enabled'::text, 'mute'::text])))
);


ALTER TABLE public.user_chat OWNER TO i9;

--
-- Name: user_chat_id_seq; Type: SEQUENCE; Schema: public; Owner: i9
--

CREATE SEQUENCE public.user_chat_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_chat_id_seq OWNER TO i9;

--
-- Name: user_chat_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: i9
--

ALTER SEQUENCE public.user_chat_id_seq OWNED BY public.user_chat.id;


--
-- Name: blocked_user id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.blocked_user ALTER COLUMN id SET DEFAULT nextval('public.blocked_user_id_seq'::regclass);


--
-- Name: comment_ id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.comment_ ALTER COLUMN id SET DEFAULT nextval('public.comment_id_seq'::regclass);


--
-- Name: chat id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat ALTER COLUMN id SET DEFAULT nextval('public.chat_id_seq'::regclass);


--
-- Name: follow id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.follow ALTER COLUMN id SET DEFAULT nextval('public.follow_id_seq'::regclass);


--
-- Name: i9l_user id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.i9l_user ALTER COLUMN id SET DEFAULT nextval('public.i9l_user_id_seq'::regclass);


--
-- Name: message_ id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_ ALTER COLUMN id SET DEFAULT nextval('public.message_id_seq'::regclass);


--
-- Name: message_deletion_log id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_deletion_log ALTER COLUMN id SET DEFAULT nextval('public.message_deletion_log_id_seq'::regclass);


--
-- Name: message_reaction id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_reaction ALTER COLUMN id SET DEFAULT nextval('public.message_reaction_id_seq'::regclass);


--
-- Name: notification id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.notification ALTER COLUMN id SET DEFAULT nextval('public.notification_id_seq'::regclass);


--
-- Name: pc_hashtag id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_hashtag ALTER COLUMN id SET DEFAULT nextval('public.pc_hashtag_id_seq'::regclass);


--
-- Name: pc_mention id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_mention ALTER COLUMN id SET DEFAULT nextval('public.pc_mention_id_seq'::regclass);


--
-- Name: pc_reaction id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_reaction ALTER COLUMN id SET DEFAULT nextval('public.pc_reaction_id_seq'::regclass);


--
-- Name: post id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.post ALTER COLUMN id SET DEFAULT nextval('public.post_id_seq'::regclass);


--
-- Name: reported_message id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.reported_message ALTER COLUMN id SET DEFAULT nextval('public.reported_message_id_seq'::regclass);


--
-- Name: repost id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.repost ALTER COLUMN id SET DEFAULT nextval('public.repost_id_seq'::regclass);


--
-- Name: saved_post id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.saved_post ALTER COLUMN id SET DEFAULT nextval('public.saved_post_id_seq'::regclass);


--
-- Name: user_chat id; Type: DEFAULT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_chat ALTER COLUMN id SET DEFAULT nextval('public.user_chat_id_seq'::regclass);


--
-- Data for Name: blocked_user; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.blocked_user (id, blocking_user_id, blocked_user_id, blocked_at) FROM stdin;
\.


--
-- Data for Name: comment_; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.comment_ (id, comment_text, commenter_user_id, attachment_url, post_id, comment_id, created_at) FROM stdin;
\.


--
-- Data for Name: chat; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.chat (id, created_at, initiator_user_id, with_user_id) FROM stdin;
\.


--
-- Data for Name: follow; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.follow (id, follower_user_id, followee_user_id, follow_on) FROM stdin;
95	14	15	2024-11-09 20:22:31.197094
96	15	14	2024-11-09 20:29:23.486856
\.


--
-- Data for Name: i9l_user; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.i9l_user (id, email, username, password, name, birthday, bio, profile_pic_url, connection_status, last_active, acc_deleted, cover_pic_url) FROM stdin;
14	johnny@gmail.com	johnny	$2b$10$R7CabqQLCGO.XtmwA5YVK.qB28vF/.5nQYUNy0pchPyUkbfCyb1CS	Johnny Cage	2000-11-07	This is a second account for testing!		online	\N	f	
15	annak@gmail.com	kendrick	$2b$10$31B7ZJy7ZaSrn9dwwaNA/u0d/ZfsY59EiEyIc3VuLUFCwGEbDcuhW	Anna Kendrick	2000-11-07	This is Anna Kendrick's account!		online	\N	f	
\.


--
-- Data for Name: message_; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.message_ (id, sender_user_id, chat_id, msg_content, created_at, delivery_status, reply_to_id) FROM stdin;
\.


--
-- Data for Name: message_deletion_log; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.message_deletion_log (id, deleter_user_id, message_id, deleted_for) FROM stdin;
\.


--
-- Data for Name: message_reaction; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.message_reaction (id, message_id, reactor_user_id, reaction_code_point) FROM stdin;
\.


--
-- Data for Name: notification; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.notification (id, type, is_read, sender_user_id, receiver_user_id, via_post_id, via_comment_id, comment_created_id, created_at, reaction_code_point) FROM stdin;
291	follow	f	14	15	\N	\N	\N	2024-11-09 13:58:28.949581	\N
292	follow	f	14	15	\N	\N	\N	2024-11-09 14:07:58.27792	\N
293	follow	f	14	15	\N	\N	\N	2024-11-09 14:11:55.388571	\N
294	follow	f	14	15	\N	\N	\N	2024-11-09 14:22:43.821553	\N
295	follow	f	14	15	\N	\N	\N	2024-11-09 14:28:56.556161	\N
296	follow	f	14	15	\N	\N	\N	2024-11-09 14:29:33.249321	\N
297	follow	f	14	15	\N	\N	\N	2024-11-09 14:34:19.515045	\N
298	follow	f	14	15	\N	\N	\N	2024-11-09 19:51:17.55273	\N
299	follow	f	14	15	\N	\N	\N	2024-11-09 19:59:18.787253	\N
300	follow	f	14	15	\N	\N	\N	2024-11-09 20:18:55.89074	\N
301	follow	f	14	15	\N	\N	\N	2024-11-09 20:22:31.197094	\N
302	follow	f	15	14	\N	\N	\N	2024-11-09 20:29:23.486856	\N
\.


--
-- Data for Name: ongoing_registration; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.ongoing_registration (sid, sess, expire) FROM stdin;
\.


--
-- Data for Name: pc_hashtag; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.pc_hashtag (id, post_id, comment_id, hashtag_name) FROM stdin;
\.


--
-- Data for Name: pc_mention; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.pc_mention (id, post_id, comment_id, user_id) FROM stdin;
\.


--
-- Data for Name: pc_reaction; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.pc_reaction (id, reactor_user_id, post_id, comment_id, reaction_code_point, created_at) FROM stdin;
\.


--
-- Data for Name: post; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.post (id, user_id, media_urls, description, type, created_at) FROM stdin;
\.


--
-- Data for Name: reported_message; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.reported_message (id, reporting_user_id, reported_user_id, message_id, reason, reported_at) FROM stdin;
\.


--
-- Data for Name: repost; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.repost (id, reposter_user_id, post_id) FROM stdin;
\.


--
-- Data for Name: saved_post; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.saved_post (id, saver_user_id, post_id) FROM stdin;
\.


--
-- Data for Name: user_chat; Type: TABLE DATA; Schema: public; Owner: i9
--

COPY public.user_chat (id, user_id, chat_id, unread_messages_count, notification_mode, deleted, updated_at, partner_user_id) FROM stdin;
\.


--
-- Name: blocked_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.blocked_user_id_seq', 1, false);


--
-- Name: comment_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.comment_id_seq', 45, true);


--
-- Name: chat_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.chat_id_seq', 1, true);


--
-- Name: follow_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.follow_id_seq', 96, true);


--
-- Name: i9l_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.i9l_user_id_seq', 15, true);


--
-- Name: message_deletion_log_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.message_deletion_log_id_seq', 2, true);


--
-- Name: message_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.message_id_seq', 7, true);


--
-- Name: message_reaction_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.message_reaction_id_seq', 3, true);


--
-- Name: notification_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.notification_id_seq', 302, true);


--
-- Name: pc_hashtag_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.pc_hashtag_id_seq', 2, true);


--
-- Name: pc_mention_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.pc_mention_id_seq', 106, true);


--
-- Name: pc_reaction_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.pc_reaction_id_seq', 37, true);


--
-- Name: post_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.post_id_seq', 10, true);


--
-- Name: reported_message_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.reported_message_id_seq', 1, false);


--
-- Name: repost_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.repost_id_seq', 15, true);


--
-- Name: saved_post_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.saved_post_id_seq', 15, true);


--
-- Name: user_chat_id_seq; Type: SEQUENCE SET; Schema: public; Owner: i9
--

SELECT pg_catalog.setval('public.user_chat_id_seq', 2, true);


--
-- Name: blocked_user BlockedUser_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.blocked_user
    ADD CONSTRAINT "BlockedUser_pkey" PRIMARY KEY (id);


--
-- Name: comment_ Comment_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.comment_
    ADD CONSTRAINT "Comment_pkey" PRIMARY KEY (id);


--
-- Name: chat Chat_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat
    ADD CONSTRAINT "Chat_pkey" PRIMARY KEY (id);


--
-- Name: follow FollowAction_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT "FollowAction_pkey" PRIMARY KEY (id);


--
-- Name: pc_hashtag Hashtag_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_hashtag
    ADD CONSTRAINT "Hashtag_pkey" PRIMARY KEY (id);


--
-- Name: pc_mention Mention_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_mention
    ADD CONSTRAINT "Mention_pkey" PRIMARY KEY (id);


--
-- Name: message_reaction MessageReaction_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_reaction
    ADD CONSTRAINT "MessageReaction_pkey" PRIMARY KEY (id);


--
-- Name: message_ Message_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_
    ADD CONSTRAINT "Message_pkey" PRIMARY KEY (id);


--
-- Name: notification Notification_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT "Notification_pkey" PRIMARY KEY (id);


--
-- Name: pc_hashtag PostCommentHashtag_hashtag_name_comment_id_key; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_hashtag
    ADD CONSTRAINT "PostCommentHashtag_hashtag_name_comment_id_key" UNIQUE (hashtag_name, comment_id);


--
-- Name: pc_hashtag PostCommentHashtag_hashtag_name_post_id_key; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_hashtag
    ADD CONSTRAINT "PostCommentHashtag_hashtag_name_post_id_key" UNIQUE (hashtag_name, post_id);


--
-- Name: pc_mention PostCommentMention_user_id_comment_id_key; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_mention
    ADD CONSTRAINT "PostCommentMention_user_id_comment_id_key" UNIQUE (user_id, comment_id);


--
-- Name: pc_mention PostCommentMention_user_id_post_id_key; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_mention
    ADD CONSTRAINT "PostCommentMention_user_id_post_id_key" UNIQUE (user_id, post_id);


--
-- Name: pc_reaction Reaction_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_reaction
    ADD CONSTRAINT "Reaction_pkey" PRIMARY KEY (id);


--
-- Name: reported_message ReportedMesssage_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.reported_message
    ADD CONSTRAINT "ReportedMesssage_pkey" PRIMARY KEY (id);


--
-- Name: repost Repost_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.repost
    ADD CONSTRAINT "Repost_pkey" PRIMARY KEY (id);


--
-- Name: saved_post SavedPost_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.saved_post
    ADD CONSTRAINT "SavedPost_pkey" PRIMARY KEY (id);


--
-- Name: user_chat UserChat_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_chat
    ADD CONSTRAINT "UserChat_pkey" PRIMARY KEY (id);


--
-- Name: i9l_user User_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.i9l_user
    ADD CONSTRAINT "User_pkey" PRIMARY KEY (id);


--
-- Name: blocked_user blocking_is_once_per_user; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.blocked_user
    ADD CONSTRAINT blocking_is_once_per_user UNIQUE (blocking_user_id, blocked_user_id);


--
-- Name: follow follow_is_once; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT follow_is_once UNIQUE (follower_user_id, followee_user_id);


--
-- Name: message_deletion_log message_deletion_log_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_deletion_log
    ADD CONSTRAINT message_deletion_log_pkey PRIMARY KEY (id);


--
-- Name: pc_reaction one_comment_reaction_per_user; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_reaction
    ADD CONSTRAINT one_comment_reaction_per_user UNIQUE (reactor_user_id, comment_id);


--
-- Name: pc_reaction one_post_reaction_per_user; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_reaction
    ADD CONSTRAINT one_post_reaction_per_user UNIQUE (reactor_user_id, post_id);


--
-- Name: saved_post one_post_save_per_user; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.saved_post
    ADD CONSTRAINT one_post_save_per_user UNIQUE (saver_user_id, post_id);


--
-- Name: post post_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT post_pkey PRIMARY KEY (id);


--
-- Name: repost repost_once; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.repost
    ADD CONSTRAINT repost_once UNIQUE (reposter_user_id, post_id);


--
-- Name: saved_post save_once; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.saved_post
    ADD CONSTRAINT save_once UNIQUE (saver_user_id, post_id);


--
-- Name: ongoing_registration session_pkey; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.ongoing_registration
    ADD CONSTRAINT session_pkey PRIMARY KEY (sid);


--
-- Name: i9l_user unique_email; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.i9l_user
    ADD CONSTRAINT unique_email UNIQUE (email);


--
-- Name: i9l_user unique_username; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.i9l_user
    ADD CONSTRAINT unique_username UNIQUE (username);


--
-- Name: user_chat userX_to_chatX_once; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_chat
    ADD CONSTRAINT "userX_to_chatX_once" UNIQUE (user_id, chat_id);


--
-- Name: message_reaction user_reacts_once; Type: CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_reaction
    ADD CONSTRAINT user_reacts_once UNIQUE (message_id, reactor_user_id);


--
-- Name: IDX_session_expire; Type: INDEX; Schema: public; Owner: i9
--

CREATE INDEX "IDX_session_expire" ON public.ongoing_registration USING btree (expire);


--
-- Name: blocked_user blocked_user; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.blocked_user
    ADD CONSTRAINT blocked_user FOREIGN KEY (blocked_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: blocked_user blocking_user; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.blocked_user
    ADD CONSTRAINT blocking_user FOREIGN KEY (blocking_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: comment_ comment_by; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.comment_
    ADD CONSTRAINT comment_by FOREIGN KEY (commenter_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: comment_ comment_commented_on; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.comment_
    ADD CONSTRAINT comment_commented_on FOREIGN KEY (comment_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: notification comment_created; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT comment_created FOREIGN KEY (comment_created_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: pc_mention comment_mentioned_in; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_mention
    ADD CONSTRAINT comment_mentioned_in FOREIGN KEY (comment_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: pc_reaction comment_reacted_to; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_reaction
    ADD CONSTRAINT comment_reacted_to FOREIGN KEY (comment_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: chat chat_initiator_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat
    ADD CONSTRAINT chat_initiator_user_id_fkey FOREIGN KEY (initiator_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: chat chat_with_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.chat
    ADD CONSTRAINT chat_with_user_id_fkey FOREIGN KEY (with_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: user_chat chat_participant; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_chat
    ADD CONSTRAINT chat_participant FOREIGN KEY (user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: follow followed_user; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT followed_user FOREIGN KEY (followee_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: follow follower_user; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT follower_user FOREIGN KEY (follower_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: pc_hashtag hashtaged_comment; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_hashtag
    ADD CONSTRAINT hashtaged_comment FOREIGN KEY (comment_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: pc_hashtag hashtaged_post; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_hashtag
    ADD CONSTRAINT hashtaged_post FOREIGN KEY (post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: message_deletion_log message_deletion_log_deleter_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_deletion_log
    ADD CONSTRAINT message_deletion_log_deleter_user_id_fkey FOREIGN KEY (deleter_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: message_deletion_log message_deletion_log_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_deletion_log
    ADD CONSTRAINT message_deletion_log_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.message_(id) ON DELETE CASCADE;


--
-- Name: message_reaction message_reacted_to; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_reaction
    ADD CONSTRAINT message_reacted_to FOREIGN KEY (message_id) REFERENCES public.message_(id) ON DELETE CASCADE;


--
-- Name: reported_message message_reported; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.reported_message
    ADD CONSTRAINT message_reported FOREIGN KEY (message_id) REFERENCES public.message_(id) ON DELETE CASCADE;


--
-- Name: message_ message_sender; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_
    ADD CONSTRAINT message_sender FOREIGN KEY (sender_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: notification notification_receiver; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT notification_receiver FOREIGN KEY (receiver_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: notification notification_sender; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT notification_sender FOREIGN KEY (sender_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: message_ owner_chat; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_
    ADD CONSTRAINT owner_chat FOREIGN KEY (chat_id) REFERENCES public.chat(id) ON DELETE CASCADE;


--
-- Name: user_chat owner_chat; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_chat
    ADD CONSTRAINT owner_chat FOREIGN KEY (chat_id) REFERENCES public.chat(id) ON DELETE CASCADE;


--
-- Name: comment_ post_commented_on; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.comment_
    ADD CONSTRAINT post_commented_on FOREIGN KEY (post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: pc_mention post_mentioned_in; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_mention
    ADD CONSTRAINT post_mentioned_in FOREIGN KEY (post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: pc_reaction post_reacted_to; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_reaction
    ADD CONSTRAINT post_reacted_to FOREIGN KEY (post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: saved_post post_saver; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.saved_post
    ADD CONSTRAINT post_saver FOREIGN KEY (saver_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: post posted_by; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT posted_by FOREIGN KEY (user_id) REFERENCES public.i9l_user(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: pc_reaction reaction_by; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_reaction
    ADD CONSTRAINT reaction_by FOREIGN KEY (reactor_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: message_reaction reactor_user; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_reaction
    ADD CONSTRAINT reactor_user FOREIGN KEY (reactor_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: message_ replied_message; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.message_
    ADD CONSTRAINT replied_message FOREIGN KEY (reply_to_id) REFERENCES public.message_(id) ON DELETE CASCADE;


--
-- Name: reported_message reported_user; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.reported_message
    ADD CONSTRAINT reported_user FOREIGN KEY (reported_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: reported_message reporting_user; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.reported_message
    ADD CONSTRAINT reporting_user FOREIGN KEY (reporting_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: repost reposted_post; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.repost
    ADD CONSTRAINT reposted_post FOREIGN KEY (post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: repost reposter; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.repost
    ADD CONSTRAINT reposter FOREIGN KEY (reposter_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: saved_post saved_post; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.saved_post
    ADD CONSTRAINT saved_post FOREIGN KEY (post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: notification through_comment; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT through_comment FOREIGN KEY (via_comment_id) REFERENCES public.comment_(id) ON DELETE CASCADE;


--
-- Name: notification through_post; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT through_post FOREIGN KEY (via_post_id) REFERENCES public.post(id) ON DELETE CASCADE;


--
-- Name: user_chat user_chat_partner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.user_chat
    ADD CONSTRAINT user_chat_partner_user_id_fkey FOREIGN KEY (partner_user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- Name: pc_mention user_mentioned; Type: FK CONSTRAINT; Schema: public; Owner: i9
--

ALTER TABLE ONLY public.pc_mention
    ADD CONSTRAINT user_mentioned FOREIGN KEY (user_id) REFERENCES public.i9l_user(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

