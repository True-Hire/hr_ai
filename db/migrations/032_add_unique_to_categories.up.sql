ALTER TABLE main_category ADD CONSTRAINT uni_main_category_name UNIQUE (name);
ALTER TABLE sub_category ADD CONSTRAINT uni_sub_category_name UNIQUE (name);
ALTER TABLE technologies ADD CONSTRAINT uni_technologies_name UNIQUE (name);
