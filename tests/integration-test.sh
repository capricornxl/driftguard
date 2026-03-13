#!/bin/bash

# DriftGuard 集成测试脚本
# 用法：./tests/integration-test.sh

set -e

BASE_URL="http://localhost:8080"
AGENT_ID="integration-test-agent"

echo "========================================"
echo "DriftGuard 集成测试"
echo "========================================"
echo ""

# 1. 健康检查
echo "1️⃣  健康检查..."
HEALTH=$(curl -s "$BASE_URL/health")
if echo "$HEALTH" | grep -q "healthy"; then
    echo "   ✅ 服务健康：$HEALTH"
else
    echo "   ❌ 服务不健康：$HEALTH"
    exit 1
fi

# 2. 推送测试数据
echo ""
echo "2️⃣  推送测试数据 (20 条交互)..."
for i in $(seq 1 20); do
    LATENCY=$((200 + i * 5))
    curl -s -X POST "$BASE_URL/api/v1/interactions" \
      -H "Content-Type: application/json" \
      -d "{\"agent_id\":\"$AGENT_ID\",\"session_id\":\"session-$i\",\"input\":\"测试输入 $i\",\"output\":\"测试输出 $i\",\"latency_ms\":$LATENCY,\"tokens_in\":10,\"tokens_out\":20}" > /dev/null
done
echo "   ✅ 已推送 20 条交互数据"

# 3. 等待数据处理
echo ""
echo "3️⃣  等待数据处理..."
sleep 2

# 4. 评估健康度
echo ""
echo "4️⃣  评估健康度..."
EVAL_RESULT=$(curl -s -X POST "$BASE_URL/api/v1/agents/$AGENT_ID/evaluate")
if echo "$EVAL_RESULT" | grep -q "score"; then
    SCORE=$(echo "$EVAL_RESULT" | grep -o '"score":[0-9.]*' | cut -d: -f2)
    LEVEL=$(echo "$EVAL_RESULT" | grep -o '"level":"[^"]*"' | cut -d'"' -f4)
    echo "   ✅ 健康度评分：$SCORE (等级：$LEVEL)"
else
    echo "   ⚠️  评估结果：$EVAL_RESULT"
fi

# 5. 退化检测
echo ""
echo "5️⃣  退化检测..."
DETECT_RESULT=$(curl -s "$BASE_URL/api/v1/agents/$AGENT_ID/detect")
if echo "$DETECT_RESULT" | grep -q "is_degraded"; then
    DEGRADED=$(echo "$DETECT_RESULT" | grep -o '"is_degraded":[a-z]*' | cut -d: -f2)
    TREND=$(echo "$DETECT_RESULT" | grep -o '"trend":"[^"]*"' | cut -d'"' -f4)
    echo "   ✅ 退化状态：$DEGRADED (趋势：$TREND)"
else
    echo "   ⚠️  检测结果：$DETECT_RESULT"
fi

# 6. 查询告警
echo ""
echo "6️⃣  查询告警..."
ALERTS=$(curl -s "$BASE_URL/api/v1/alerts")
ALERT_COUNT=$(echo "$ALERTS" | grep -o '"total":[0-9]*' | cut -d: -f2)
echo "   ✅ 告警数量：$ALERT_COUNT"

# 7. 统计信息
echo ""
echo "7️⃣  统计信息..."
STATS=$(curl -s "$BASE_URL/api/v1/stats")
if echo "$STATS" | grep -q "collector"; then
    echo "   ✅ 统计信息获取成功"
else
    echo "   ⚠️  统计信息：$STATS"
fi

# 8. Prometheus 检查
echo ""
echo "8️⃣  Prometheus 检查..."
PROM_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9090/api/v1/targets)
if [ "$PROM_STATUS" = "200" ]; then
    echo "   ✅ Prometheus 正常运行"
else
    echo "   ⚠️  Prometheus 状态码：$PROM_STATUS"
fi

# 9. Grafana 检查
echo ""
echo "9️⃣  Grafana 检查..."
GRAFANA_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/health)
if [ "$GRAFANA_STATUS" = "200" ]; then
    echo "   ✅ Grafana 正常运行"
else
    echo "   ⚠️  Grafana 状态码：$GRAFANA_STATUS"
fi

echo ""
echo "========================================"
echo "集成测试完成！"
echo "========================================"
echo ""
echo "访问地址:"
echo "  - API:      http://localhost:8080"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana:    http://localhost:3000 (admin/driftguard)"
echo ""
