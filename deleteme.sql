-- this file contains the schema for a basic social media app

-- account tracks a single user.
-- an account can own multiple profiles, and have multiple wallets
CREATE TABLE accounts (
    id UUID PRIMARY KEY
);

insert into accounts (id) values ('d33fd73c-e81c-463c-9773-5947fc20d23d'::uuid);

-- profiles tracks the public-facing information about a user
CREATE TABLE profiles (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL CHECK (length(username) > 0 AND length(username) <= 50),
    age INT NOT NULL CHECK (age >= 0),
    bio TEXT NOT NULL CHECK (length(bio) <= 500),
    account_id UUID NOT NULL REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE CASCADE
);

insert into profiles (id, username, age, bio, account_id) values
('afdc0da2-c49b-4952-9cc9-027f4c7640bc'::uuid, 'satoshi', 30, 'I am the creator of Bitcoin', 'd33fd73c-e81c-463c-9773-5947fc20d23d'::uuid),
('e1a227bc-2e5f-4603-962d-b32f1d085973'::uuid, 'vitalik', 27, 'I am the creator of Ethereum', 'd33fd73c-e81c-463c-9773-5947fc20d23d'::uuid);

-- wallets tracks the wallet information for a user
CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    address TEXT UNIQUE NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE CASCADE
);

insert into wallets (id, address, account_id) values
('c656107c-7049-4d4a-af2f-f093f9fa6e7c'::uuid, '0xAfFDC06cF34aFD7D5801A13d48C92AD39609901D', 'd33fd73c-e81c-463c-9773-5947fc20d23d'::uuid);

-- posts tracks the posts made by users
-- posts can be threaded, like a comment on a post
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    content TEXT NOT NULL CHECK (length(content) <= 500),
    created_at INT NOT NULL,
    author_id UUID NOT NULL REFERENCES profiles(id) ON UPDATE CASCADE ON DELETE CASCADE,
    parent_id UUID REFERENCES posts(id) ON UPDATE CASCADE ON DELETE CASCADE
);

insert into posts (id, content, created_at, author_id, parent_id) values
('f1b1b1b1-1b1b-1b1b-1b1b-1b1b1b1b1b1b'::uuid, 'Hello, world!', 1630000000, 'afdc0da2-c49b-4952-9cc9-027f4c7640bc'::uuid, null),
('f1b1b1b1-1b1b-1b1b-1b1b-1b1b1b1b1b1c'::uuid, 'Hello, world! 2', 1630000001, 'e1a227bc-2e5f-4603-962d-b32f1d085973'::uuid, 'f1b1b1b1-1b1b-1b1b-1b1b-1b1b1b1b1b1b'::uuid),
('f1b1b1b1-1b1b-1b1b-1b1b-1b1b1b1b1b1d'::uuid, 'Hello, world! 3', 1630000002, 'afdc0da2-c49b-4952-9cc9-027f4c7640bc'::uuid, 'f1b1b1b1-1b1b-1b1b-1b1b-1b1b1b1b1b1b'::uuid);

-- likes tracks the likes made by users
CREATE TABLE likes (
    post_id UUID NOT NULL REFERENCES posts(id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES profiles(id) ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (post_id, user_id)
);

-- friends tracks the relationships between users
CREATE TABLE friends (
    user_id UUID NOT NULL REFERENCES profiles(id) ON UPDATE CASCADE ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES profiles(id) ON UPDATE CASCADE ON DELETE CASCADE,
    PRIMARY KEY (user_id, friend_id)
);


-- register_account creates a new account with the wallet address of the caller attached to it
CREATE ACTION register_account() public returns (UUID) {
    -- uuid_generate_kwil is a special Kwil function that uses a fixed namespace
    -- to generate a deterministic UUID based on the transaction ID
    $id = uuid_generate_kwil(@txid||'account'); 
    INSERT INTO accounts (id) VALUES ($id);
    INSERT INTO wallets (id, address, account_id) VALUES (uuid_generate_kwil(@txid||'wallet'), @caller, $id);

    return $id;
};

-- register_wallet creates a new wallet for the account of the caller
CREATE ACTION register_wallet($address text) public {
    $account_id := account_id(@caller);
    if $account_id is null {
        error('Account does not exist');
    }

    INSERT INTO wallets (id, address, account_id) VALUES (uuid_generate_kwil(@txid||'wallet'), $address, $account_id);
};

-- account_id gets the account id of a wallet address.
-- If none exists, it returns null.
CREATE ACTION account_id($address TEXT) public view returns (UUID) {
    for $row in SELECT account_id FROM wallets WHERE address = $address {
        return $row.account_id;
    }

    return null;
};

-- remove_wallet removes a wallet from the account of the caller
CREATE ACTION remove_wallet($address TEXT) public {
    $account_id := account_id(@caller);
    if $account_id is null {
        error('Account does not exist');
    }

    DELETE FROM wallets WHERE address = $address AND account_id = $account_id;
};

-- create_profile creates a new profile for the account of the caller.
-- If the account does not exist, it will be created.
CREATE ACTION create_profile($username TEXT, $age INT, $bio TEXT) public {
    $account_id := account_id(@caller);
    if $account_id is null {
        $account_id = register_account();
    }

    INSERT INTO profiles (id, username, age, bio, account_id) VALUES (
        uuid_generate_kwil(@txid||'profile'||$username),
        $username, $age, $bio, $account_id
    );
};

-- owns_profile checks if the wallet address owns the profile name
CREATE ACTION owns_profile($address TEXT, $username TEXT) public view returns (BOOL) {
    for $row in SELECT 1 FROM profiles p JOIN wallets w ON p.account_id = w.account_id WHERE w.address = $address AND p.username = $username {
        return true;
    }

    return false;
};

-- create_post creates a new post for the specified profile
-- If the caller does not own the profile, an error is thrown
CREATE ACTION create_post($username TEXT, $content TEXT, $parent_id UUID) public {
    if !owns_profile(@caller, $username) {
        error('You do not own this profile');
    }

    INSERT INTO posts (id, content, created_at, author_id, parent_id) VALUES (
        uuid_generate_kwil(@txid||'post'),
        $content,
        @height,
        (SELECT id FROM profiles WHERE username = $username),
        $parent_id
    );
};

-- like_post likes a post for the specified profile
-- If the caller does not own the profile, an error is thrown
CREATE ACTION like_post($username TEXT, $post_id UUID) public {
    if !owns_profile(@caller, $username) {
        error('You do not own this profile');
    }

    INSERT INTO likes (post_id, user_id) VALUES (
        $post_id,
        (SELECT id FROM profiles WHERE username = $username)
    );
};

-- get_likes returns the number of likes for a post
CREATE ACTION get_likes($post_id UUID) public view returns (INT) {
    return SELECT COUNT(*) FROM likes WHERE post_id = $post_id;
};

-- add_friend adds a friend relationship between two profiles
-- If the caller does not own the profile, an error is thrown
CREATE ACTION add_friend($username TEXT, $friend_username TEXT) public {
    if !owns_profile(@caller, $username) {
        error('You do not own this profile');
    }

    INSERT INTO friends (user_id, friend_id) VALUES (
        (SELECT id FROM profiles WHERE username = $username),
        (SELECT id FROM profiles WHERE username = $friend_username)
    );
};

-- get_friends returns the list of friends for the specified profile
CREATE ACTION get_friends($username TEXT) public view returns TABLE(username TEXT) {
    return SELECT pr.username AS username
    FROM friends f
    JOIN profiles pr
    ON f.friend_id = pr.id
    JOIN profiles pr2
    ON f.user_id = pr2.id
    WHERE pr2.username = $username;
};

-- get_posts returns the list of posts for the specified profile
CREATE ACTION get_posts($username TEXT) public view returns table(post_id UUID, content TEXT, created_at INT, likes INT) {
    return WITH likes AS (
        SELECT post_id, COUNT(*) as likes FROM likes GROUP BY post_id
    )
    SELECT p.id, p.content, p,created_at, COALESCE(l.likes, 0) as likes FROM posts p
    LEFT JOIN likes l
    ON p.id = l.post_id
    JOIN profiles pr
    ON p.author_id = pr.id
    WHERE pr.username = $username
    ORDER BY p.created_at DESC;
};

-- get_thread gets a post, its comments, and the comments' comments, recursively
CREATE ACTION get_thread($post_id UUID, $max_depth int) public view returns table(post_id UUID, content TEXT, created_at INT, author TEXT, likes INT, children UUID[], depth INT) {
    if $max_depth < 0 {
        error('max_depth must be greater than or equal to 0');
    }
    if $max_depth > 5 {
        error('max_depth must be less than or equal to 5');
    }
    if $max_depth is null {
        $max_depth = 2;
    }
    
    return WITH RECURSIVE children AS (
        SELECT id, content, created_at, author_id, parent_id, 0 as depth
        FROM posts WHERE parent_id = $post_id
        UNION ALL
        SELECT p.id, p.content, p.created_at, p.author_id, p.parent_id, c.depth + 1
        FROM posts p 
        JOIN children c
        ON p.parent_id = c.id
        WHERE c.depth < $max_depth
    ), like_counts AS (
        SELECT post_id, COUNT(*) as likes FROM likes GROUP BY post_id
    )
    SELECT c.id as post_id, c.content, c.created_at, pr.username as author, COALESCE(l.likes, 0) as likes, ARRAY_AGG(c.id) as children, c.depth
    FROM children c
    LEFT JOIN like_counts l
    ON c.id = l.post_id
    JOIN profiles pr
    ON c.author_id = pr.id
    GROUP BY c.id, c.content, c.created_at, pr.username, l.likes, c.depth
    ORDER BY c.depth;
};

-- update_profile updates the profile information for the account
-- that is tied to $username
CREATE ACTION update_profile($old_username TEXT, $new_username TEXT, $age INT, $bio TEXT) public {
    if !owns_profile(@caller, $old_username) {
        error('You do not own this profile');
    }

    UPDATE profiles
    SET username = $new_username, age = $age, bio = $bio
    WHERE username = $old_username;
};

-- delete_profile deletes the profile information for the account
-- that is tied to $username
CREATE ACTION delete_profile($username TEXT) public {
    if !owns_profile(@caller, $username) {
        error('You do not own this profile');
    }

    DELETE FROM profiles WHERE username = $username;
};