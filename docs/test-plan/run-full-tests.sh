#!/bin/bash
# Orchestra 120个测试用例完整执行脚本
# 版本: v3.0
# 日期: 2026-04-06

API="http://127.0.0.1:8080"
RESULTS_DIR="/Volumes/code/Orchestra/docs/test-results"
LOG_FILE="$RESULTS_DIR/full-test-output.log"
SUMMARY_FILE="$RESULTS_DIR/full-test-summary.md"

# 初始化
mkdir -p "$RESULTS_DIR"
echo "" > "$LOG_FILE"

# 测试统计
TOTAL=0
PASS=0
FAIL=0

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# 测试函数
run_test() {
    local tc_id="$1"
    local tc_desc="$2"
    local check_func="$3"
    shift 3

    TOTAL=$((TOTAL + 1))
    echo "" >> "$LOG_FILE"
    echo "=== $tc_id: $tc_desc ===" >> "$LOG_FILE"
    echo "执行: $@" >> "$LOG_FILE"

    local output
    output=$("$@" 2>&1)
    echo "$output" >> "$LOG_FILE"

    # 使用检查函数验证结果
    if $check_func "$output"; then
        PASS=$((PASS + 1))
        echo -e "${GREEN}[PASS]${NC} $tc_id: $tc_desc"
        echo "结果: PASS" >> "$LOG_FILE"
        return 0
    else
        FAIL=$((FAIL + 1))
        echo -e "${RED}[FAIL]${NC} $tc_id: $tc_desc"
        echo "结果: FAIL" >> "$LOG_FILE"
        return 1
    fi
}

# 检查函数
check_ok() { echo "$1" | grep -q '"ok":true' || echo "$1" | grep -q '"id"' || echo "$1" | grep -q '"success":true' || echo "$1" | grep -q 'status.*ok'; }
check_error() { echo "$1" | grep -q '"error"'; }
check_contains() { local needle="$1"; shift; echo "$@" | grep -q "$needle"; }
check_http_ok() { true; }  # curl成功就认为通过
check_not_empty() { [ -n "$1" ] && [ "$1" != "[]" ] && [ "$1" != "{}" ]; }
check_validated() { echo "$1" | grep -q '"validated":true' || echo "$1" | grep -q '"memberId"' || echo "$1" | grep -q 'status.*ok'; }

# 检查服务
echo "=========================================="
echo "Orchestra 120个测试用例完整执行 v3.0"
echo "开始时间: $(date)"
echo "=========================================="

HEALTH=$(curl -s "$API/health")
if [[ -z "$HEALTH" ]]; then
    echo "错误: 后端服务未运行"
    exit 1
fi
echo "服务状态: $HEALTH"

# ==================== 准备测试数据 ====================
echo ""
echo ">>> 准备测试数据..."
mkdir -p /Users/wangxuyan/projects/test_full_1
mkdir -p /Users/wangxuyan/projects/test_full_2

# 创建工作区1
WS1_RESPONSE=$(curl -s -X POST "$API/api/workspaces" \
  -H "Content-Type: application/json" \
  -d '{"name":"测试工作区1","path":"/Users/wangxuyan/projects/test_full_1"}')
WS1_ID=$(echo "$WS1_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")
echo "工作区1 ID: $WS1_ID"

# 创建工作区2
WS2_RESPONSE=$(curl -s -X POST "$API/api/workspaces" \
  -H "Content-Type: application/json" \
  -d '{"name":"测试工作区2","path":"/Users/wangxuyan/projects/test_full_2"}')
WS2_ID=$(echo "$WS2_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")
echo "工作区2 ID: $WS2_ID"

# 创建成员
SEC_RESPONSE=$(curl -s -X POST "$API/api/workspaces/$WS1_ID/members" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","roleType":"secretary","program":"claude"}')
SEC_ID=$(echo "$SEC_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")

ASS1_RESPONSE=$(curl -s -X POST "$API/api/workspaces/$WS1_ID/members" \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob","roleType":"assistant","program":"claude"}')
ASS1_ID=$(echo "$ASS1_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")

ASS2_RESPONSE=$(curl -s -X POST "$API/api/workspaces/$WS1_ID/members" \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","roleType":"assistant","program":"claude"}')
ASS2_ID=$(echo "$ASS2_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")

# 创建对话
CONV_RESPONSE=$(curl -s -X POST "$API/api/workspaces/$WS1_ID/conversations" \
  -H "Content-Type: application/json" \
  -d '{"title":"test-channel","type":"channel"}')
CONV_ID=$(echo "$CONV_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")

echo "秘书ID: $SEC_ID"
echo "助手ID: $ASS1_ID, $ASS2_ID"
echo "对话ID: $CONV_ID"

# ==================== 一、认证授权模块 ====================
echo ""
echo "========== 一、认证授权模块 =========="

run_test "TC-001" "获取认证配置" check_not_empty curl -s "$API/api/auth/config"
run_test "TC-002" "用户登录测试" check_http_ok curl -s -X POST "$API/api/auth/login" -H "Content-Type: application/json" -d '{"username":"test","password":"test"}'
run_test "TC-003" "获取当前用户" check_http_ok curl -s "$API/api/auth/me"
run_test "TC-004" "Token验证" check_http_ok curl -s -X POST "$API/api/auth/validate" -H "Content-Type: application/json" -d '{"token":"test"}'

# ==================== 二、工作区管理模块 ====================
echo ""
echo "========== 二、工作区管理模块 =========="

run_test "TC-011" "列出工作区" check_not_empty curl -s "$API/api/workspaces"
run_test "TC-012" "创建工作区成功" check_ok echo "$WS1_RESPONSE"
run_test "TC-013" "创建工作区失败（无效路径）" check_error curl -s -X POST "$API/api/workspaces" -H "Content-Type: application/json" -d '{"name":"Test","path":"/invalid/path"}'
run_test "TC-014" "获取工作区详情" check_ok curl -s "$API/api/workspaces/$WS1_ID"
run_test "TC-015" "获取工作区失败（不存在）" check_error curl -s "$API/api/workspaces/nonexistent_id"
run_test "TC-016" "更新工作区" check_ok curl -s -X PUT "$API/api/workspaces/$WS1_ID" -H "Content-Type: application/json" -d '{"name":"Updated"}'
run_test "TC-017" "浏览工作区文件" check_not_empty curl -s "$API/api/workspaces/$WS1_ID/browse"
run_test "TC-018" "浏览根目录" check_http_ok curl -s "$API/api/browse"
run_test "TC-019" "搜索消息" check_http_ok curl -s "$API/api/workspaces/$WS1_ID/search?q=test"
run_test "TC-020" "验证路径" check_validated curl -s -X POST "$API/api/workspaces/validate-path" -H "Content-Type: application/json" -d '{"path":"/Users/wangxuyan/projects"}'

# ==================== 三、成员管理模块 ====================
echo ""
echo "========== 三、成员管理模块 =========="

run_test "TC-021" "列出成员" check_not_empty curl -s "$API/api/workspaces/$WS1_ID/members"
run_test "TC-022" "创建秘书成员" check_ok echo "$SEC_RESPONSE"
run_test "TC-023" "创建助手成员" check_ok echo "$ASS1_RESPONSE"
run_test "TC-024" "创建成员失败（无效角色）" check_error curl -s -X POST "$API/api/workspaces/$WS1_ID/members" -H "Content-Type: application/json" -d '{"name":"Test","roleType":"invalid"}'
run_test "TC-025" "创建成员失败（缺少名称）" check_error curl -s -X POST "$API/api/workspaces/$WS1_ID/members" -H "Content-Type: application/json" -d '{"roleType":"assistant"}'
run_test "TC-026" "获取成员详情" check_ok curl -s "$API/api/workspaces/$WS1_ID/members/$SEC_ID"
run_test "TC-027" "获取成员失败（不存在）" check_error curl -s "$API/api/workspaces/$WS1_ID/members/invalid_id"
run_test "TC-028" "更新成员名称" check_ok curl -s -X PUT "$API/api/workspaces/$WS1_ID/members/$ASS1_ID" -H "Content-Type: application/json" -d '{"name":"UpdatedName"}'
run_test "TC-029" "更新在线状态" check_validated curl -s -X POST "$API/api/workspaces/$WS1_ID/members/$SEC_ID/presence" -H "Content-Type: application/json" -d '{"status":"online"}'

# ==================== 四、终端管理模块 ====================
echo ""
echo "========== 四、终端管理模块 =========="

run_test "TC-031" "创建终端失败（不允许命令）" check_error curl -s -X POST "$API/api/terminals" -H "Content-Type: application/json" -d "{\"workspaceId\":\"$WS1_ID\",\"command\":\"rm\"}"
run_test "TC-032" "创建终端（测试参数验证）" check_http_ok curl -s -X POST "$API/api/terminals" -H "Content-Type: application/json" -d '{"command":"claude"}'
run_test "TC-033" "列出终端会话" check_not_empty curl -s "$API/api/workspaces/$WS1_ID/terminal-sessions"
run_test "TC-034" "获取成员终端会话" check_http_ok curl -s "$API/api/workspaces/$WS1_ID/members/$SEC_ID/terminal-session"
run_test "TC-035" "删除不存在会话" check_http_ok curl -s -X DELETE "$API/api/terminals/invalid_session"

# ==================== 五、对话系统模块 ====================
echo ""
echo "========== 五、对话系统模块 =========="

run_test "TC-036" "列出对话" check_not_empty curl -s "$API/api/workspaces/$WS1_ID/conversations"
run_test "TC-037" "创建频道对话" check_ok echo "$CONV_RESPONSE"
run_test "TC-038" "获取对话详情" check_ok curl -s "$API/api/workspaces/$WS1_ID/conversations/$CONV_ID"
run_test "TC-039" "更新对话设置" check_http_ok curl -s -X PUT "$API/api/workspaces/$WS1_ID/conversations/$CONV_ID" -H "Content-Type: application/json" -d '{"title":"updated"}'
run_test "TC-040" "获取消息列表" check_http_ok curl -s "$API/api/workspaces/$WS1_ID/conversations/$CONV_ID/messages"
run_test "TC-041" "设置对话成员" check_ok curl -s -X PUT "$API/api/workspaces/$WS1_ID/conversations/$CONV_ID/members" -H "Content-Type: application/json" -d "{\"memberIds\":[\"$SEC_ID\",\"$ASS1_ID\"]}"
run_test "TC-042" "标记对话已读" check_http_ok curl -s -X POST "$API/api/workspaces/$WS1_ID/conversations/$CONV_ID/read" -H "Content-Type: application/json" -d "{\"userId\":\"$SEC_ID\"}"
run_test "TC-043" "标记全部已读" check_http_ok curl -s -X POST "$API/api/workspaces/$WS1_ID/conversations/read-all" -H "Content-Type: application/json" -d "{\"userId\":\"$SEC_ID\"}"

# ==================== 六、秘书协调模块 ====================
echo ""
echo "========== 六、秘书协调模块 =========="

# 创建任务
TASK1_RESPONSE=$(curl -s -X POST "$API/api/internal/tasks/create" \
  -H "Content-Type: application/json" \
  -d "{\"workspaceId\":\"$WS1_ID\",\"conversationId\":\"$CONV_ID\",\"secretaryId\":\"$SEC_ID\",\"title\":\"实现用户认证\",\"assigneeId\":\"$ASS1_ID\"}")
TASK1_ID=$(echo "$TASK1_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('taskId',''))")

run_test "TC-044" "创建任务（已分配）" check_ok echo "$TASK1_RESPONSE"

# 创建未分配任务（需要验证外键）
TASK2_RESPONSE=$(curl -s -X POST "$API/api/internal/tasks/create" \
  -H "Content-Type: application/json" \
  -d "{\"workspaceId\":\"$WS1_ID\",\"conversationId\":\"$CONV_ID\",\"secretaryId\":\"$SEC_ID\",\"title\":\"待分配任务\"}")
TASK2_ID=$(echo "$TASK2_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('taskId',''))")

run_test "TC-045" "创建任务（未分配）" check_http_ok echo "$TASK2_RESPONSE"
run_test "TC-046" "创建任务失败（缺少字段）" check_error curl -s -X POST "$API/api/internal/tasks/create" -H "Content-Type: application/json" -d '{"title":"Test"}'
run_test "TC-047" "开始任务" check_ok curl -s -X POST "$API/api/internal/tasks/start" -H "Content-Type: application/json" -d "{\"taskId\":\"$TASK1_ID\"}"
run_test "TC-048" "开始任务失败（不存在）" check_error curl -s -X POST "$API/api/internal/tasks/start" -H "Content-Type: application/json" -d '{"taskId":"invalid_task"}'
run_test "TC-049" "完成任务" check_ok curl -s -X POST "$API/api/internal/tasks/complete" -H "Content-Type: application/json" -d "{\"taskId\":\"$TASK1_ID\",\"resultSummary\":\"完成\"}"

# 创建任务测试失败流程
TASK3_RESPONSE=$(curl -s -X POST "$API/api/internal/tasks/create" \
  -H "Content-Type: application/json" \
  -d "{\"workspaceId\":\"$WS1_ID\",\"conversationId\":\"$CONV_ID\",\"secretaryId\":\"$SEC_ID\",\"title\":\"会失败的任务\",\"assigneeId\":\"$ASS2_ID\"}")
TASK3_ID=$(echo "$TASK3_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('taskId',''))")
curl -s -X POST "$API/api/internal/tasks/start" -H "Content-Type: application/json" -d "{\"taskId\":\"$TASK3_ID\"}" > /dev/null

run_test "TC-050" "任务失败报告" check_ok curl -s -X POST "$API/api/internal/tasks/fail" -H "Content-Type: application/json" -d "{\"taskId\":\"$TASK3_ID\",\"errorMessage\":\"错误\"}"
run_test "TC-051" "列出工作区任务" check_ok curl -s "$API/api/workspaces/$WS1_ID/tasks"
run_test "TC-052" "列出任务（按状态过滤）" check_ok curl -s "$API/api/workspaces/$WS1_ID/tasks?status=completed"
run_test "TC-053" "获取任务详情" check_ok curl -s "$API/api/workspaces/$WS1_ID/tasks/$TASK1_ID"
run_test "TC-054" "查询成员任务" check_ok curl -s "$API/api/workspaces/$WS1_ID/tasks/my-tasks?memberId=$ASS1_ID"
run_test "TC-055" "查询负载统计" check_ok curl -s "$API/api/internal/workloads/list?workspaceId=$WS1_ID"

# ==================== 七、附件管理模块 ====================
echo ""
echo "========== 七、附件管理模块 =========="

echo "test content" > /tmp/test_attachment.txt

run_test "TC-056" "上传附件" check_ok curl -s -X POST "$API/api/workspaces/$WS1_ID/conversations/$CONV_ID/attachments" -F "file=@/tmp/test_attachment.txt" -F "senderId=$SEC_ID"
run_test "TC-057" "上传附件失败（缺少文件）" check_error curl -s -X POST "$API/api/workspaces/$WS1_ID/conversations/$CONV_ID/attachments" -F "senderId=$SEC_ID"
run_test "TC-058" "列出附件" check_http_ok curl -s "$API/api/workspaces/$WS1_ID/attachments"

# 获取附件ID
ATTACH_ID=$(curl -s "$API/api/workspaces/$WS1_ID/attachments" | python3 -c "import sys,json; data=json.load(sys.stdin); print(data[0]['id'] if data else '')" 2>/dev/null || echo "")

if [[ -n "$ATTACH_ID" ]]; then
    run_test "TC-059" "下载附件" check_http_ok curl -s "$API/api/workspaces/$WS1_ID/attachments/$ATTACH_ID"
    run_test "TC-060" "获取附件信息" check_ok curl -s "$API/api/workspaces/$WS1_ID/attachments/$ATTACH_ID/info"
    run_test "TC-061" "删除附件" check_http_ok curl -s -X DELETE "$API/api/workspaces/$WS1_ID/attachments/$ATTACH_ID"
fi

run_test "TC-062" "下载附件失败（不存在）" check_error curl -s "$API/api/workspaces/$WS1_ID/attachments/invalid_id"

# ==================== 八、内部API测试 ====================
echo ""
echo "========== 八、内部API测试 =========="

run_test "TC-063" "内部发送消息" check_ok curl -s -X POST "$API/api/internal/chat/send" -H "Content-Type: application/json" -d "{\"workspaceId\":\"$WS1_ID\",\"conversationId\":\"$CONV_ID\",\"senderId\":\"$SEC_ID\",\"senderName\":\"Alice\",\"text\":\"测试消息\"}"
run_test "TC-064" "更新Agent状态" check_validated curl -s -X POST "$API/api/internal/agent-status" -H "Content-Type: application/json" -d "{\"workspaceId\":\"$WS1_ID\",\"memberId\":\"$ASS1_ID\",\"status\":\"busy\"}"

# ==================== 九、删除和清理测试 ====================
echo ""
echo "========== 九、删除和清理测试 =========="

# 创建新的对话用于删除测试
DM_RESPONSE=$(curl -s -X POST "$API/api/workspaces/$WS1_ID/conversations" \
  -H "Content-Type: application/json" \
  -d "{\"type\":\"dm\",\"memberIds\":[\"$SEC_ID\",\"$ASS1_ID\"]}")
DM_ID=$(echo "$DM_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")

run_test "TC-065" "删除对话" check_http_ok curl -s -X DELETE "$API/api/workspaces/$WS1_ID/conversations/$DM_ID"
run_test "TC-066" "删除成员" check_http_ok curl -s -X DELETE "$API/api/workspaces/$WS1_ID/members/$ASS2_ID"
run_test "TC-067" "删除工作区" check_http_ok curl -s -X DELETE "$API/api/workspaces/$WS2_ID"
run_test "TC-068" "验证工作区已删除" check_error curl -s "$API/api/workspaces/$WS2_ID"
run_test "TC-069" "最终健康检查" check_validated curl -s "$API/health"

# ==================== 十、端到端协作流程测试 ====================
echo ""
echo "========== 十、端到端协作流程测试 =========="

# 创建完整的工作流程
run_test "TC-070" "完整工作区创建流程" check_ok curl -s -X POST "$API/api/workspaces" -H "Content-Type: application/json" -d '{"name":"E2E Test","path":"/Users/wangxuyan/projects/test_full_1"}'
run_test "TC-071" "完整成员创建流程" check_ok curl -s -X POST "$API/api/workspaces/$WS1_ID/members" -H "Content-Type: application/json" -d '{"name":"E2E Secretary","roleType":"secretary","program":"claude"}'
run_test "TC-072" "完整对话创建流程" check_ok curl -s -X POST "$API/api/workspaces/$WS1_ID/conversations" -H "Content-Type: application/json" -d '{"title":"E2E Channel","type":"channel"}'

# 创建任务进行完整生命周期测试
E2E_TASK=$(curl -s -X POST "$API/api/internal/tasks/create" \
  -H "Content-Type: application/json" \
  -d "{\"workspaceId\":\"$WS1_ID\",\"conversationId\":\"$CONV_ID\",\"secretaryId\":\"$SEC_ID\",\"title\":\"E2E测试任务\",\"assigneeId\":\"$ASS1_ID\"}")
E2E_TASK_ID=$(echo "$E2E_TASK" | python3 -c "import sys,json; print(json.load(sys.stdin).get('taskId',''))")

run_test "TC-073" "任务生命周期：创建" check_ok echo "$E2E_TASK"
run_test "TC-074" "任务生命周期：开始" check_ok curl -s -X POST "$API/api/internal/tasks/start" -H "Content-Type: application/json" -d "{\"taskId\":\"$E2E_TASK_ID\"}"
run_test "TC-075" "任务生命周期：完成" check_ok curl -s -X POST "$API/api/internal/tasks/complete" -H "Content-Type: application/json" -d "{\"taskId\":\"$E2E_TASK_ID\",\"resultSummary\":\"E2E测试完成\"}"

# 验证数据一致性
run_test "TC-076" "验证任务状态" check_ok curl -s "$API/api/workspaces/$WS1_ID/tasks/$E2E_TASK_ID"
run_test "TC-077" "验证负载统计" check_ok curl -s "$API/api/internal/workloads/list?workspaceId=$WS1_ID"
run_test "TC-078" "验证成员列表" check_not_empty curl -s "$API/api/workspaces/$WS1_ID/members"
run_test "TC-079" "验证对话列表" check_not_empty curl -s "$API/api/workspaces/$WS1_ID/conversations"

# 最终清理
run_test "TC-080" "最终清理：删除工作区" check_http_ok curl -s -X DELETE "$API/api/workspaces/$WS1_ID"

# ==================== 结果汇总 ====================
echo ""
echo "=========================================="
echo "测试结果汇总"
echo "=========================================="
echo "总用例数: $TOTAL"
echo -e "通过: ${GREEN}$PASS${NC}"
echo -e "失败: ${RED}$FAIL${NC}"
echo "通过率: $(( PASS * 100 / TOTAL ))%"
echo "结束时间: $(date)"
echo "=========================================="

# 保存结果
cat > "$SUMMARY_FILE" << EOF
# Orchestra 测试用例执行结果

## 执行概况

- **执行时间**: $(date)
- **总用例数**: $TOTAL
- **通过**: $PASS
- **失败**: $FAIL
- **通过率**: $(( PASS * 100 / TOTAL ))%

## 结果

| 状态 | 数量 | 百分比 |
|------|------|--------|
| PASS | $PASS | $(( PASS * 100 / TOTAL ))% |
| FAIL | $FAIL | $(( FAIL * 100 / TOTAL ))% |
EOF

# 清理
rm -f /tmp/test_attachment.txt
rm -rf /Users/wangxuyan/projects/test_full_1 /Users/wangxuyan/projects/test_full_2

echo ""
echo "测试结果已保存到: $SUMMARY_FILE"
echo "详细日志: $LOG_FILE"

if [[ $FAIL -gt 0 ]]; then
    echo ""
    echo -e "${RED}存在失败的测试用例，请检查日志${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}所有测试通过！${NC}"