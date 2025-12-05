### 整体约定

- **存储类型**：全部用 Nakama Storage（`collection + key + user_id + value(JSON)`）。  
- **命名约定**：
  - **玩家私有数据**：`user_id = 玩家ID`。
  - **全局/配置数据**（静态或运营）：`user_id = ""` 或专门用 `system` 用户。

---

### 1. 角色与团队

- **`collection = "characters"`**
  - **key**：`"main"`（每个玩家一个主角）
  - **user_id**：玩家 ID
  - **value 示例**：

    ```json
    {
      "name": "苏明远",
      "gender": "male",
      "appearance": {
        "body_type": "standard",
        "face_preset": 1
      },
      "faction": "晋商",
      "background": "落榜秀才",
      "reputation": {
        "level": "无名",
        "value": 0
      },
      "level": 1,
      "stats": { "str": 8, "agi": 8, "vit": 9, "int": 15 },
      "hp": 106,
      "def": 13,
      "equipment": {
        "head": "青布头巾",
        "top": "麻布短褐",
        "bottom": "棉布裈裤",
        "shoes": "编草鞋",
        "weapon": "青竹手杖"
      },
      "skills": {
        "识货": 1,
        "议价": 1,
        "打探": 0,
        "攻击": 1,
        "重劈": 0,
        "破甲": 0,
        "包扎": 0,
        "治疗": 0,
        "手术": 0
      },
      "created_at": 1710000000000
    }
    ```

- **`collection = "team"`**
  - **key**：`"members"`
  - **user_id**：玩家 ID
  - **value**：

    ```json
    {
      "formation": "鱼鳞阵",
      "members": [
        {
          "id": "member_qin_tieshan",
          "role_type": "护卫",
          "name": "秦铁山",
          "quality": "普通",
          "quality_coeff": 0.5,
          "stats": { "str": 8, "agi": 6, "vit": 9, "int": 5 },
          "skills": { "重劈": 1 }
        },
        {
          "id": "member_su_mingyuan",
          "role_type": "账房",
          "name": "苏明远",
          "quality": "普通",
          "quality_coeff": 0.5,
          "stats": { "str": 4, "agi": 5, "vit": 5, "int": 9 },
          "skills": {}
        }
      ]
    }
    ```

---

### 2. 背包与货币 / 道具

- **`collection = "inventory"`**
  - **key**：`"currency"`
  - **user_id**：玩家 ID
  - **value**：

    ```json
    {
      "silver": 1200,
      "gold": 0
    }
    ```

  - **key**：`"items"`
  - **user_id**：玩家 ID
  - **value**：

    ```json
    {
      "consumables": {
        "金疮药": 5,
        "麻沸散": 1,
        "烟雾弹": 2
      },
      "goods": [
        { "id": "good_mi_xiaomai", "name": "小麦", "category": "米粮", "quantity": 100 },
        { "id": "good_jiu_fenjiu", "name": "汾酒", "category": "酒类", "quantity": 10 }
      ]
    }
    ```

---

### 3. 世界 / 路线 / 跑商状态

- **`collection = "world"`**
  - **key**：`"location"`
  - **user_id**：玩家 ID
  - **value**（仅在节点）：

    ```json
    {
      "type": "node",
      "node_id": "taiyuan_fu"
    }
    ```

  - **value**（在路上）：

    ```json
    {
      "type": "travel",
      "from": "taiyuan_fu",
      "to": "yamen_guan",
      "start_time": 1710000100000,
      "duration_sec": 1800,
      "risk_factor": 0.3
    }
    ```

- **`collection = "world_config"`（全局静态）**
  - **key**：`"nodes"`，`user_id = ""`
  - **value**（示例，简化）：

    ```json
    {
      "taiyuan_fu": { "name": "太原府", "type": "city" },
      "shiling_guan": { "name": "石岭关", "type": "gate" },
      "yamen_guan": { "name": "雁门关", "type": "gate" },
      "datong_fu": { "name": "大同府", "type": "city" }
    }
    ```

  - **key**：`"routes"`，`user_id = ""`
  - **value**：

    ```json
    [
      { "from": "taiyuan_fu", "to": "shiling_guan", "distance": 10, "base_duration_sec": 600, "base_risk": 0.0 },
      { "from": "shiling_guan", "to": "yamen_guan", "distance": 20, "base_duration_sec": 1200, "base_risk": 0.3 },
      { "from": "yamen_guan", "to": "datong_fu", "distance": 25, "base_duration_sec": 1500, "base_risk": 0.5 }
    ]
    ```

- **`collection = "world_events"`（全局或按服运营）**
  - **key**：`"route_modifiers"`
  - **user_id**：`""`
  - **value**：

    ```json
    {
      "shiling_guan-yamen_guan": {
        "risk_delta": -0.25,
        "expires_at": 1710003600000,
        "reason": "云中寨被扫荡"
      },
      "yamen_guan-datong_fu": {
        "risk_delta": -0.3,
        "has_support_call": true,
        "reason": "雁门关商栈建立"
      }
    }
    ```

---

### 4. 战斗与敌人配置

- **`collection = "combat_config"`（全局静态）**
  - **key**：`"enemy_templates"`，`user_id = ""`
  - **value**：

    ```json
    {
      "流民":  { "str": 8,  "agi": 10, "vit": 9,  "hp": 106, "def": 30, "weapon_damage": 8 },
      "山匪":  { "str": 12, "agi": 10, "vit": 11, "hp": 134, "def": 35, "weapon_damage": 10 },
      "山匪头目": { "str": 14, "agi": 11, "vit": 12, "hp": 148, "def": 40, "weapon_damage": 12 },
      "马贼":  { "str": 11, "agi": 14, "vit": 10, "hp": 122, "def": 40, "weapon_damage": 11 }
    }
    ```

  - **key**：`"enemy_groups"`，`user_id = ""`
  - **value**：

    ```json
    {
      "流民团伙": [
        { "template": "流民", "count": 3 }
      ],
      "山匪小队": [
        { "template": "山匪头目", "count": 1 },
        { "template": "山匪", "count": 2 }
      ],
      "马贼突袭队": [
        { "template": "马贼", "count": 2 }
      ]
    }
    ```

  - **key**：`"formations"`，`user_id = ""`
  - **value**：

    ```json
    {
      "鱼鳞阵": { "damage_pct": 0.10, "def_pct": -0.05, "agi_flat": 0 },
      "方圆阵": { "damage_pct": 0.0,  "def_pct": 0.10, "agi_flat": 0 },
      "雁行阵": { "damage_pct": 0.0,  "def_pct": 0.0,  "agi_flat": 5 }
    }
    ```

- **（可选）`collection = "combat_logs"`**
  - **key**：战斗 ID（如 UUID）
  - **user_id**：玩家 ID（或 `""` 做审计）
  - **value**：完整战斗回放数据（回合日志等），Demo 阶段可省略。

---

### 5. 商品与城市经济（跑商）

- **`collection = "goods_config"`（全局静态）**
  - **key**：`"categories"`，`user_id = ""`
  - **value**：品类列表（米粮/布帛/皮货/...）

  - **key**：`"items"`，`user_id = ""`
  - **value**：

    ```json
    {
      "good_mi_xiaomai": {
        "name": "小麦",
        "category": "米粮",
        "base_price": 60,
        "rarity": 1.0
      },
      "good_jiu_fenjiu": {
        "name": "汾酒",
        "category": "酒类",
        "base_price": 1200,
        "rarity": 0.5,
        "stars": 3
      }
    }
    ```

- **`collection = "city_state"`**
  - **key**：城市 ID（如 `"taiyuan_fu"`）
  - **user_id**：`""`
  - **value**：

    ```json
    {
      "dev_level": 50,       // D
      "preferences": {       // D_local
        "米粮": 0.5,
        "酒类": 0.8,
        "皮货": 0.2
      },
      "events": {
        "DemandShock": { "good_jiu_fenjiu": 1.2 },
        "SupplyShock": { "good_pi_pao": 0.8 }
      }
    }
    ```

- **`collection = "goods_global"`**
  - **key**：`"global_factors"`
  - **user_id**：`""`
  - **value**：

    ```json
    {
      "good_mi_xiaomai": { "GlobalFactor": 1.0, "recent_volume": 100000 },
      "good_jiu_fenjiu": { "GlobalFactor": 1.3, "recent_volume": 5000 }
    }
    ```

- **`collection = "player_reputation"`**
  - **key**：城市 ID（如 `"taiyuan_fu"`）
  - **user_id**：玩家 ID
  - **value**：

    ```json
    {
      "reputation_level": 3,
      "reputation_value": 250
    }
    ```

---

这份 schema 是“草案级别”：字段可以再精简或拆分，但整体结构已经对齐你文档里的所有主要对象（角色、团队、世界位置、战斗、跑商）。  
如果你希望下一步落到具体实现，我可以按这个 schema 帮你写一组对口的 Lua RPC 骨架（例如 `characters.get/create`、`world.travel_start/finish`、`trade.buy/sell`）。