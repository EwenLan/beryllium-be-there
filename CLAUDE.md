# 项目概览
beryllium-be-there -- 一个课堂上使用的签到管理程序

## 技术栈
- 语言：TypeScript
- 运行时：Go / Node.js

## 项目结构
beryllium-server/ # Go语言开发的后端服务
beryllium-manage/ # 签到管理系统的管理页面
beryllium-signin/ # 签到管理系统的签到页面

## 代码风格与规范
- 类型：严格规范代码中定义的变量、结构体类型，避免强制转换、类型断言、Go语言中使用interface{}、TypeScript使用any类型
- 导入顺序：标准库>第三方库>自定义库
- 注释：每个函数、方法在开头处简要描述功能，关键代码部分添加注释描述具体业务功能

## 架构与设计约束
- 安全：涉及密码的部分遵循当前主流安全设计，禁止使用DES、md5等不安全算法
- 设计：这是一个演示项目，只需要实现必要的业务能力
- 架构：这是一个全栈项目，包括beryllium-server服务后端，beryllium-manage前端管理页面，beryllium-singin前端签到页面
  - beryllium-server：