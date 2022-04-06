--- DROP COLUMN
ALTER TABLE `fileserver`.`fileserver` DROP `digest`;
ALTER TABLE `fileserver`.`fileserver` DROP `upload_type`;
ALTER TABLE `fileserver`.`fileserver` DROP `store_name`;
ALTER TABLE `fileserver`.`fileserver` DROP `upload_id`;
ALTER TABLE `fileserver`.`fileserver` DROP `number`;
ALTER TABLE `fileserver`.`fileserver` DROP `file_name`;

--- ADD COLUMN
ALTER TABLE `fileserver`.`fileserver` ADD COLUMN `id` VARCHAR(36) FIRST;

UPDATE `fileserver`.`fileserver` set `id` = UUID();
DELETE FROM `fileserver`.`fileserver` WHERE `id` NOT IN (
    SELECT fs.minid FROM ( SELECT MIN(id) AS minid FROM `fileserver`.`fileserver` GROUP BY `path`) fs 
);

--- ADD PRIMARY KEY and UNIQUE KEY
ALTER TABLE `fileserver`.`fileserver` ADD PRIMARY KEY(`id`);
ALTER TABLE `fileserver`.`fileserver` ADD UNIQUE UQE_PATH (`path`);