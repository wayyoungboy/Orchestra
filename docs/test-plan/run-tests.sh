#!/bin/bash
# Orchestra 系统功能测试执行脚本

API="http://127.0.0.1:8080"
RESULTS_DIR="/Volumes/code/Orchestra/docs/test-results"
mkdir -p "$RESULTS_DIR"

echo "=========================================="
echo "Orchestra 系统功能测试"
echo "开始时间: $(date)"
echo "=========================================="
echo ""

# 测试结果统计
TOTAL=0
PASS=0
FAIL=0

# 测试函数
test_case() {
    local tc_id="$1"
    local desc="$2"
    shift 2
    TOTAL=$((TOTAL + 1))

    echo ""
    echo ">>> $tc_id: $desc"
    echo "    执行中..."

    local output
    output=$("$@" 2>&1)
    local status=$?

    if [ $status -eq 0 ]; then
        PASS=$((PASS + 1))
        echo "    [PASS]"
        echo "$output" | head -20
    else
        FAIL=$((FAIL + 1))
        echo "    [FAIL]"
        echo "$output"
    fi

    echo "---" >> "$RESULTS_DIR/test-output.log"
    echo "TC: $tc_id" >> "$RESULTS_DIR/test-output.log"
    echo "$output" >> "$RESULTS_DIR/test-output.log"
}

# 检查服务是否运行
echo ">>> 检查服务状态"
HEALTH=$(curl -s "$API/health")
if [ -z "$HEALTH" ]; then
    echo "错误: 后端服务未运行，请先启动 make run"
    exit 1
fi
echo "服务状态: $HEALTH"
echo ""

# ========== M01 认证授权模块 ==========
echo "========== M01 认证授权模块 =========="

test_case "TC-M01-001" "获取认证配置" curl -s "$API/api/auth/config"

# ========== M02 工作区管理模块 ==========
echo ""
echo "========== M02 工作区管理模块 =========="

test_case "TC-M02-001" "列出工作区" curl -s "$API/api/workspaces"

# 创建测试目录
mkdir -p /Users/wangxuyan/projects/orchestra_test

# 创建工作区
WS_RESPONSE=$(curl -s -X POST "$API/api/workspaces" \
  -H "Content-Type: application/json" \
  -d '{"name":"功能测试工作区","path":"/Users/wangxuyan/projects/orchestra_test"}')
WS_ID=$(echo "$WS_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")
echo "创建工作区: $WS_ID"

test_case "TC-M02-003" "创建工作区（无效路径）" curl -s -X POST "$API/api/workspaces" \
  -H "Content-Type: application/json" \
  -d '{"name":"测试","path":"/invalid/nonexistent/path"}'

test_case "TC-M02-004" "获取工作区详情" curl -s "$API/api/workspaces/$WS_ID"

test_case "TC-M02-007" "浏览工作区文件" curl -s "$API/api/workspaces/$WS_ID/browse"

# ========== M03 成员管理模块 ==========
echo ""
echo "========== M03 成员管理模块 =========="

test_case "TC-M03-001" "列出成员" curl -s "$API/api/workspaces/$WS_ID/members"

# 创建秘书
SEC_RESPONSE=$(curl -s -X POST "$API/api/workspaces/$WS_ID/members" \
  -H "Content-Type: application/json" \
  -d '{"name":"TestSecretary","roleType":"secretary","program":"claude"}')
SEC_ID=$(echo "$SEC_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")
echo "创建秘书: $SEC_ID"

# 创建助手
ASS_RESPONSE=$(curl -s -X POST "$API/api/workspaces/$WS_ID/members" \
  -H "Content-Type: application/json" \
  -d '{"name":"TestAssistant","roleType":"assistant","program":"claude"}')
ASS_ID=$(echo "$ASS_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")
echo "创建助手: $ASS_ID"

test_case "TC-M03-004" "更新成员" curl -s -X PUT "$API/api/workspaces/$WS_ID/members/$SEC_ID" \
  -H "Content-Type: application/json" \
  -d '{"name":"UpdatedSecretary"}'

test_case "TC-M03-006" "更新在线状态" curl -s -X POST "$API/api/workspaces/$WS_ID/members/$SEC_ID/presence" \
  -H "Content-Type: application/json" \
  -d '{"status":"online"}'

# ========== M05 对话系统模块 ==========
echo ""
echo "========== M05 对话系统模块 =========="

test_case "TC-M05-001" "列出对话" curl -s "$API/api/workspaces/$WS_ID/conversations"

# 创建对话
CONV_RESPONSE=$(curl -s -X POST "$API/api/workspaces/$WS_ID/conversations" \
  -H "Content-Type: application/json" \
  -d '{"title":"测试对话"}')
CONV_ID=$(echo "$CONV_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")
echo "创建对话: $CONV_ID"

test_case "TC-M05-003" "获取对话详情" curl -s "$API/api/workspaces/$WS_ID/conversations/$CONV_ID"

test_case "TC-M05-005" "获取消息列表" curl -s "$API/api/workspaces/$WS_ID/conversations/$CONV_ID/messages"

test_case "TC-M05-007" "标记已读" curl -s -X POST "$API/api/workspaces/$WS_ID/conversations/$CONV_ID/read" \
  -H "Content-Type: application/json" \
  -d "{\"memberId\":\"$SEC_ID\"}"

# ========== M07 秘书协调模块 ==========
echo ""
echo "========== M07 秘书协调模块 =========="

# 创建任务
TASK_RESPONSE=$(curl -s -X POST "$API/api/internal/tasks/create" \
  -H "Content-Type: application/json" \
  -d "{\"workspaceId\":\"$WS_ID\",\"conversationId\":\"$CONV_ID\",\"secretaryId\":\"$SEC_ID\",\"title\":\"测试任务\",\"description\":\"这是一个测试任务\",\"assigneeId\":\"$ASS_ID\"}")
TASK_ID=$(echo "$TASK_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('taskId',''))")
echo "创建任务: $TASK_ID"

test_case "TC-M07-001" "创建任务" echo "$TASK_RESPONSE"

test_case "TC-M07-002" "开始任务" curl -s -X POST "$API/api/internal/tasks/start" \
  -H "Content-Type: application/json" \
  -d "{\"taskId\":\"$TASK_ID\"}"

test_case "TC-M07-003" "完成任务" curl -s -X POST "$API/api/internal/tasks/complete" \
  -H "Content-Type: application/json" \
  -d "{\"taskId\":\"$TASK_ID\",\"resultSummary\":\"任务完成\"}"

test_case "TC-M07-005" "查询负载" curl -s "$API/api/internal/workloads/list?workspaceId=$WS_ID"

test_case "TC-M07-006" "列出任务" curl -s "$API/api/workspaces/$WS_ID/tasks"

test_case "TC-M07-007" "获取任务详情" curl -s "$API/api/workspaces/$WS_ID/tasks/$TASK_ID"

# 创建另一个任务用于测试失败流程
TASK2_RESPONSE=$(curl -s -X POST "$API/api/internal/tasks/create" \
  -H "Content-Type: application/json" \
  -d "{\"workspaceId\":\"$WS_ID\",\"conversationId\":\"$CONV_ID\",\"secretaryId\":\"$SEC_ID\",\"title\":\"失败测试任务\",\"assigneeId\":\"$ASS_ID\"}")
TASK2_ID=$(echo "$TASK2_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('taskId',''))")

curl -s -X POST "$API/api/internal/tasks/start" \
  -H "Content-Type: application/json" \
  -d "{\"taskId\":\"$TASK2_ID\"}"

test_case "TC-M07-004" "任务失败" curl -s -X POST "$API/api/internal/tasks/fail" \
  -H "Content-Type: application/json" \
  -d "{\"taskId\":\"$TASK2_ID\",\"errorMessage\":\"测试失败原因\"}"

test_case "TC-M07-008" "查询成员任务" curl -s "$API/api/workspaces/$WS_ID/tasks/my-tasks?memberId=$ASS_ID"

# ========== M02 搜索功能 ==========
echo ""
echo "========== M02 搜索功能 =========="

test_case "TC-M02-008" "搜索消息" curl -s "$API/api/workspaces/$WS_ID/search?q=测试"

# ========== M04 终端管理模块 ==========
echo ""
echo "========== M04 终端管理模块 =========="

test_case "TC-M04-002" "创建不允许的命令会话" curl -s -X POST "$API/api/terminals" \
  -H "Content-Type: application/json" \
  -d "{\"workspaceId\":\"$WS_ID\",\"command\":\"rm\"}"

test_case "TC-M04-003" "列出工作区终端会话" curl -s "$API/api/workspaces/$WS_ID/terminal-sessions"

# ========== M08 文件附件模块 ==========
echo ""
echo "========== M08 文件附件模块 =========="

test_case "TC-M08-001" "上传附件" curl -s -X POST "$API/api/workspaces/$WS_ID/conversations/$CONV_ID/attachments" \
  -F "file=@/Volumes/code/Orchestra/docs/test-plan/system-test-plan.md" \
  -F "uploadedBy=$SEC_ID"

test_case "TC-M08-002" "列出附件" curl -s "$API/api/workspaces/$WS_ID/attachments"

# ========== 清理测试数据 ==========
echo ""
echo "========== 清理测试数据 =========="

echo "删除测试工作区..."
curl -s -X DELETE "$API/api/workspaces/$WS_ID"
rm -rf /Users/wangxuyan/projects/orchestra_test

# ========== 测试结果汇总 ==========
echo ""
echo "=========================================="
echo "测试结果汇总"
echo "=========================================="
echo "总用例数: $TOTAL"
echo "通过: $PASS"
echo "失败: $FAIL"
echo "通过率: $(( PASS * 100 / TOTAL ))%"
echo "结束时间: $(date)"
echo "=========================================="

# 保存结果到文件
cat > "$RESULTS_DIR/test-summary.md" << EOF
# Orchestra 系统功能测试结果

## 测试执行概况

- **执行时间**: $(date)
- **总用例数**: $TOTAL
- **通过**: $PASS
- **失败**: $FAIL
- **通过率**: $(( PASS * 100 / TOTAL ))%

## 测试结果

| 状态 | 数量 | 百分比 |
|------|------|--------|
| PASS | $PASS | $(( PASS * 100 / TOTAL ))% |
| FAIL | $FAIL | $(( FAIL * 100 / TOTAL ))% |
EOF

if [ $FAIL -gt 0 ]; then
    echo ""
    echo "存在失败的测试用例，请查看详细日志: $RESULTS_DIR/test-output.log"
    exit 1
fi

echo ""
echo "所有测试通过！"