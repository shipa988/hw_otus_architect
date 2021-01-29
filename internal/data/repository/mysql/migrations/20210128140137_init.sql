-- +goose Up
CREATE TABLE Users
(
    Id INT(11) PRIMARY KEY AUTO_INCREMENT,
    Login VARCHAR(255) NOT NULL UNIQUE,
    PassHash VARCHAR(255) NOT NULL
);
CREATE TABLE Profiles
(
    Id INT(11)  PRIMARY KEY,
    Name     VARCHAR(255),
    SurName  VARCHAR(255),
    Age      INT,
    Gen      VARCHAR(20),
    Interest VARCHAR(1024),
    City     VARCHAR(255),
    FOREIGN KEY (Id) REFERENCES Users(Id) ON DELETE CASCADE
);


CREATE TABLE Friends
(
    UserId INT(11)  NOT NULL,
    FriendId INT(11)  NOT NULL,
    PRIMARY KEY (UserId, FriendId),
    FOREIGN KEY (UserId) REFERENCES Users(id) ON DELETE CASCADE,
    FOREIGN KEY (FriendId) REFERENCES Users(id) ON DELETE CASCADE
);

CREATE TABLE Seanses
(
    Id INT(11) PRIMARY KEY AUTO_INCREMENT,
    UserId INT(11)  NOT NULL,
    Uuid varchar(36)  NOT NULL,
    FOREIGN KEY (UserId) REFERENCES Users(id) ON DELETE CASCADE
);
-- +goose Down
DROP table public.events;
