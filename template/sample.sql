CREATE TABLE student (
 stu_id varchar(10),
 stu_name varchar(20)
);

CREATE TABLE course (
  course_id varchar(10),
  course_name varchar(20)
);

CREATE TABLE module (
 stu_id varchar(10),
 course_id varchar(10),
 updated varchar(20)
);

INSERT INTO student VALUES
('0001','first'),
('0002','second');

INSERT INTO course VALUES
('0010', 'courseM1'),
('0011', 'courseM2'),
('0012', 'courseM3'),
('0013', 'courseM4');

INSERT INTO module VALUES
('0001','0010',''),
('0001','0011',''),
('0001','0012',''),
('0001','0013',''),
('0002','0010',''),
('0002','0011','');


/* 
select item.course_name, usr.stu_name, updated FROM student usr 
   INNER JOIN course item ON module.stu_id = usr.stu_id 
   INNER JOIN module ON module.course_id = item.course_id 
   WHERE usr.stu_name='second';
*/
