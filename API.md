# 剪纸工坊课程预约系统 - RESTful API 文档

## 基础信息

- Base URL: `/api`
- 响应格式: JSON
- 字符编码: UTF-8

## 通用响应格式

```json
{
  "data": {},
  "message": "操作成功",
  "error": "错误描述"
}
```

---

## 课程管理

### GET /api/courses

获取课程列表

**查询参数:**

| 参数  | 类型   | 必填 | 说明                          |
| ----- | ------ | ---- | ----------------------------- |
| date  | string | 否   | 按日期筛选，格式 YYYY-MM-DD   |
| month | string | 否   | 按月份筛选，格式 YYYY-MM      |

**响应示例:**

```json
{
  "data": [
    {
      "id": 1,
      "title": "剪纸入门体验课",
      "date": "2026-06-10",
      "time_slot": "09:00-11:00",
      "capacity": 15,
      "booked": 12,
      "created_at": "2026-06-05T10:00:00Z",
      "updated_at": "2026-06-05T10:00:00Z"
    }
  ]
}
```

### GET /api/courses/:id

获取单个课程详情

**路径参数:**

| 参数 | 类型  | 必填 | 说明   |
| ---- | ----- | ---- | ------ |
| id   | int64 | 是   | 课程ID |

### POST /api/courses

创建课程

**请求体:**

```json
{
  "title": "剪纸入门体验课",
  "date": "2026-06-10",
  "time_slot": "09:00-11:00",
  "capacity": 15
}
```

| 字段     | 类型   | 必填 | 说明         |
| -------- | ------ | ---- | ------------ |
| title    | string | 是   | 课程名称     |
| date     | string | 是   | 日期 YYYY-MM-DD |
| time_slot| string | 是   | 时段         |
| capacity | int    | 是   | 名额上限 ≥1  |

**响应:** `201 Created`

```json
{
  "data": { "id": 1 },
  "message": "课程创建成功"
}
```

### PUT /api/courses/:id

更新课程信息

**请求体:** 同创建课程

**业务规则:** 名额不能少于已预约人数

### DELETE /api/courses/:id

删除课程

**业务规则:** 有预约的课程不允许删除

---

## 预约管理

### POST /api/bookings

创建预约（并发安全，使用数据库行锁）

**请求体:**

```json
{
  "course_id": 1,
  "user_name": "张三",
  "user_phone": "13800138000"
}
```

| 字段       | 类型   | 必填 | 说明   |
| ---------- | ------ | ---- | ------ |
| course_id  | int64  | 是   | 课程ID |
| user_name  | string | 是   | 姓名   |
| user_phone | string | 是   | 手机号 |

**成功响应:** `201 Created`

```json
{
  "data": {
    "id": 1,
    "course_title": "剪纸入门体验课",
    "date": "2026-06-10",
    "time_slot": "09:00-11:00",
    "user_name": "张三",
    "remaining": 2
  },
  "message": "预约成功！"
}
```

**错误响应:**

- `409 Conflict` - 名额已满 / 已预约过该课程
- `404 Not Found` - 课程不存在

**并发安全说明:**

预约接口使用 `BEGIN IMMEDIATE` 事务 + `SELECT ... FOR UPDATE` 行锁，确保在高并发下名额不会超卖：

1. 开启 `IMMEDIATE` 事务获取写锁
2. `SELECT FOR UPDATE` 锁定课程行
3. 检查名额是否已满
4. 检查是否重复预约
5. 插入预约记录
6. 提交事务释放锁

### GET /api/bookings/my

查询个人预约记录

**查询参数:**

| 参数  | 类型   | 必填 | 说明   |
| ----- | ------ | ---- | ------ |
| phone | string | 是   | 手机号 |

**响应示例:**

```json
{
  "data": [
    {
      "id": 1,
      "course_id": 1,
      "user_name": "张三",
      "user_phone": "13800138000",
      "status": "booked",
      "created_at": "2026-06-05T10:30:00Z",
      "cancelled_at": null,
      "course_title": "剪纸入门体验课",
      "course_date": "2026-06-10",
      "course_slot": "09:00-11:00"
    }
  ]
}
```

### DELETE /api/bookings/:id

取消预约

**业务规则:** 取消后名额实时释放，已取消的不能重复取消

**响应:**

```json
{
  "message": "预约已取消，名额已释放"
}
```

---

## 课程预约详情

### GET /api/courses/:id/bookings

获取某课程的预约名单

**响应:** 返回该课程所有预约记录（含已取消）

### GET /api/courses/:id/availability

查询课程余量

**响应:**

```json
{
  "data": {
    "course_id": 1,
    "capacity": 15,
    "booked": 12,
    "available": 3,
    "is_full": false
  }
}
```

---

## 排期日历

### GET /api/courses/calendar

获取日历视图数据

**查询参数:**

| 参数  | 类型   | 必填 | 说明                      |
| ----- | ------ | ---- | ------------------------- |
| month | string | 否   | 月份筛选，格式 YYYY-MM    |

**响应:**

```json
{
  "data": [
    {
      "id": 1,
      "title": "剪纸入门体验课",
      "date": "2026-06-10",
      "time_slot": "09:00-11:00",
      "capacity": 15,
      "booked": 15,
      "is_full": true
    }
  ]
}
```

---

## 统计

### GET /api/stats

获取所有课程的到课率统计

**响应:**

```json
{
  "data": [
    {
      "course_id": 1,
      "course_title": "剪纸入门体验课",
      "course_date": "2026-06-10",
      "course_slot": "09:00-11:00",
      "capacity": 15,
      "booked": 14,
      "attended": 12,
      "attend_rate": 85.7
    }
  ]
}
```

---

## 到课标记

### PUT /api/attendance/:id

标记预约为已到课

**路径参数:** `id` 为预约记录ID

**业务规则:** 只有状态为 `booked` 的记录可以标记
