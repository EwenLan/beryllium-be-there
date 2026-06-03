# 开发需求
beryllium-be-there是一个在课堂上使用的签到管理程序。在beryllium-server运行后在8080端口上启动http服务器，服务器提供以下功能：
1. 提供http服务器功能：
  - `http://hostname:8080/` 展示beryllium-manage提供的管理页面
  - `http://hostname:8080/signin/{class-id}` 展示对应课堂的签到页面，并下发创建课堂时生成的公钥
  - `POST http://hostname:8080/siginin/{class-id}` Body部分携带如下的JSON格式，校验对应学生的密码是否正确，如果正确将对应用户的签到状态置为已签到，并返回对应的HTTP状态码。
```TypeScript
interface SignIn {
    "account": string,
    "password": string,
}
```
2. 提供数据库功能：由于是演示项目，目前只需要通过JSON文件实现数据库功能

显示beryllium-manage提供的管理页面，管理员输入用户名和密码登录管理系统。在管理系统内有3种功能：
1. 学生管理：通过路径/student-manage进入学生管理页面后，可以添加和删除学生，一个学生有3个属性，其中`account`是主键，添加学生后将学生信息添加到数据库中
```TypeScript
interface Student{
    "account": string,
    "name": string,
    "password": string
}
```
2. 课堂管理：通过/class-manage/进入后可以创建课堂，创建课堂后生成唯一的UUID `class_id`作为主键，通过RSA算法生成一组密钥
3. 课堂签到管理：通过/class-manage/{class-id}/进入对应课堂的签到管理页面，在页面中展示这一课堂签到用使用的二维码和地址链接`http://hostname/siginin/{class-id}`，展示所有学生的名字（即学生信息的`name`字段）和签到状态，提供导出功能：将学生和签到状态导出为csv文件

beryllium-signin提供签到页面，通过地址`http://hostname/signin/{class-id}`可以访问对应课堂的签到页面，签到页面要求输入学生的account和password。输入信息后密码RSA公钥加密输入的密码，通过`POST http://hostname:8080/siginin/{class-id}`进行签到，服务端返回成功或失败相应码展示对应提示。