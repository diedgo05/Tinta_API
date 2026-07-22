#!/bin/bash
# ═══════════════════════════════════════════════════════════════════════════
# OWASP ZAP Baseline (escaneo pasivo) contra Railway
# ═══════════════════════════════════════════════════════════════════════════

set -e

OUTPUT_DIR="$(pwd)/reports"
mkdir -p "$OUTPUT_DIR"

# URLs
declare -A SERVICES=(
    ["identity"]="https://tinta-identity.up.railway.app"
    ["communities"]="https://tinta-communities.up.railway.app"
    ["recommendations"]="https://tinta-recommendations.up.railway.app"
    ["catalog"]="https://tinta-catalog.up.railway.app"
    ["reading"]="https://tinta-reading.up.railway.app"
    ["knowledge"]="https://tinta-knowledge.up.railway.app"
    ["notifications"]="https://tinta-notifications.up.railway.app"
)

echo "🎯 OWASP ZAP Baseline (escaneo pasivo)"
echo "📁 Reportes en: $OUTPUT_DIR"
echo ""

for name in "${!SERVICES[@]}"; do
    url="${SERVICES[$name]}"
    echo "🔍 Escaneando: $name -> $url"

    docker run --rm \
        -v "$OUTPUT_DIR:/zap/wrk/:rw" \
        -t zaproxy/zap-stable \
        zap-baseline.py \
        -t "$url" \
        -r "zap-$name.html" \
        -I || true

    echo "   ✅ Reporte: reports/zap-$name.html"
    echo ""
done

echo "✅ Todos los escaneos completados"
