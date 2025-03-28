CREATE TABLE `course` (
    `sub_id` int NOT NULL,
    `course_id` int NOT NULL,
    `name` text NOT NULL,
    `teacher` text NOT NULL,
    `location` text NOT NULL,
    `date` text NOT NULL,
    `time` text NOT NULL,
    `video` text,
    `summary_status` text,
    `summary_data` text,
    PRIMARY KEY (`sub_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
