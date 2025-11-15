# Blue/Green 배포 테스트 가이드

## 개요
이 가이드는 traffic 기반으로 HTTP 200 응답이 80% 이상일 때 blue/green 배포를 테스트하는 방법을 설명합니다.

## 사전 준비사항

1. **DNS 설정** (또는 /etc/hosts)
   ```bash
   # 로컬 테스트용 - 두 개의 별도 주소 설정
   echo "<INGRESS_IP> rollouts-demo-bg-active.example.com" | sudo tee -a /etc/hosts
   echo "<INGRESS_IP> rollouts-demo-bg-preview.example.com" | sudo tee -a /etc/hosts
   ```

2. **Prometheus 메트릭 확인**
   - Prometheus가 nginx ingress 메트릭을 수집하고 있는지 확인
   - 메트릭: `nginx_ingress_controller_response_duration_seconds_count`

## 테스트 절차

### 1단계: 초기 배포 (Blue 버전)

```bash
# Rollout 상태 확인
kubectl get rollout rollouts-demo-bg -n default

# Active 서비스로 트래픽 확인
curl http://rollouts-demo-bg-active.example.com
# 또는
curl -k https://rollouts-demo-bg-active.example.com

# 현재 버전 확인 (blue 이미지)
kubectl describe rollout rollouts-demo-bg -n default
```

**예상 결과:**
- Rollout이 `Healthy` 상태
- Active 서비스가 blue 버전을 가리킴
- 기본 요청이 정상 응답 (200)

---

### 2단계: 새 버전 배포 (Green 버전)

```bash
# 새 버전(green)으로 이미지 업데이트
kubectl set image rollout/rollouts-demo-bg \
  rollouts-demo-bg=argoproj/rollouts-demo:green \
  -n default

# 또는 rollout을 직접 수정
kubectl patch rollout rollouts-demo-bg -n default --type='json' \
  -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "argoproj/rollouts-demo:green"}]'
```

**예상 결과:**
- Rollout이 새 버전(green)을 preview 서비스로 배포 시작
- Active 서비스는 여전히 blue 버전을 가리킴
- Preview 서비스가 green 버전을 가리킴

---

### 3단계: Preview 서비스 테스트 (20% 단계)

Preview 서비스로 트래픽을 보내서 테스트합니다.

```bash
# Preview 서비스 주소로 직접 요청
# 20% 비율로 트래픽 생성 (예: 10개 중 2개)
for i in {1..2}; do
  curl http://rollouts-demo-bg-preview.example.com
  sleep 1
done

# 또는 지속적으로 트래픽 생성 (분석이 3분 동안 진행되므로)
# 별도 터미널에서 실행
while true; do
  curl http://rollouts-demo-bg-preview.example.com
  sleep 5
done
```

**확인 사항:**
```bash
# Rollout 상태 확인
kubectl get rollout rollouts-demo-bg -n default

# Analysis 실행 상태 확인
kubectl get analysisrun -n default

# Preview 서비스로 라우팅되는지 확인
kubectl get svc rollouts-demo-bg-preview -n default
```

**예상 결과:**
- Preview 서비스가 green 버전을 서빙
- HTTP 200 응답이 80% 이상이어야 함
- Analysis가 성공하면 다음 단계로 진행 가능

---

### 4단계: Preview 서비스 테스트 (60% 단계)

더 많은 트래픽을 Preview 서비스로 보냅니다.

```bash
# 60% 비율로 트래픽 생성 (예: 10개 중 6개)
for i in {1..6}; do
  curl http://rollouts-demo-bg-preview.example.com
  sleep 1
done

# 지속적으로 트래픽 생성
while true; do
  curl http://rollouts-demo-bg-preview.example.com
  sleep 3
done
```

**확인 사항:**
```bash
# Analysis 메트릭 확인
kubectl describe analysisrun <analysis-run-name> -n default

# Prometheus에서 직접 확인 (선택사항)
# Prometheus UI에서 다음 쿼리 실행:
# sum(rate(nginx_ingress_controller_response_duration_seconds_count{host="rollouts-demo-bg-preview.example.com", status="200"}[3m])) / sum(rate(nginx_ingress_controller_response_duration_seconds_count{host="rollouts-demo-bg-preview.example.com"}[3m]))
```

---

### 5단계: Preview 서비스 테스트 (100% 단계)

모든 트래픽을 Preview 서비스로 보냅니다.

```bash
# 모든 요청을 Preview 서비스로
while true; do
  curl http://rollouts-demo-bg-preview.example.com
  sleep 2
done
```

**확인 사항:**
- HTTP 200 응답 비율이 80% 이상 유지되는지 확인
- 에러가 발생하지 않는지 확인

---

### 6단계: 수동 승인 및 전환

현재 설정(`autoPromotionEnabled: false`)에서는 수동으로 승인해야 합니다.

```bash
# Rollout 상태 확인
kubectl get rollout rollouts-demo-bg -n default

# Preview 버전 승인 (Active로 전환)
kubectl argo rollouts promote rollouts-demo-bg -n default

# 또는 kubectl patch 사용
kubectl patch rollout rollouts-demo-bg -n default --type='merge' \
  -p='{"status":{"promoteFull":true}}'
```

**예상 결과:**
- Active 서비스가 green 버전을 가리킴
- Preview 서비스는 제거됨 (또는 다음 배포를 위해 대기)
- 30초 후 이전 버전(blue)이 제거됨 (`scaleDownDelaySeconds: 30`)

---

### 7단계: 전환 후 확인

```bash
# Active 서비스로 기본 요청
curl http://rollouts-demo-bg-active.example.com

# Rollout 상태 확인
kubectl get rollout rollouts-demo-bg -n default

# Pod 상태 확인
kubectl get pods -l app=rollouts-demo-bg -n default

# 서비스 확인
kubectl get svc -l app=rollouts-demo-bg -n default
```

**예상 결과:**
- Active 서비스가 green 버전을 서빙
- Active 주소로 요청이 green 버전으로 라우팅
- Post-promotion analysis가 실행됨 (설정된 경우)

---

## 트러블슈팅

### 1. Preview 서비스로 트래픽이 가지 않는 경우

```bash
# Ingress 확인
kubectl get ingress -n default

# Preview Ingress 확인
kubectl describe ingress rollouts-demo-bg-preview-ingress -n default

# Preview 주소로 직접 요청 확인
curl -v http://rollouts-demo-bg-preview.example.com
```

### 2. Analysis가 실패하는 경우

```bash
# AnalysisRun 상세 확인
kubectl get analysisrun -n default
kubectl describe analysisrun <analysis-run-name> -n default

# Prometheus 연결 확인
kubectl exec -it <prometheus-pod> -n monitoring -- wget -qO- http://prometheus-kube-prometheus-prometheus.monitoring.svc:9090/api/v1/query?query=nginx_ingress_controller_response_duration_seconds_count

# 메트릭이 수집되고 있는지 확인
# Prometheus UI에서 확인하거나
curl "http://<prometheus-address>/api/v1/query?query=nginx_ingress_controller_response_duration_seconds_count{host=\"rollouts-demo-bg-preview.example.com\"}"
```

### 3. HTTP 200 응답 비율이 80% 미만인 경우

```bash
# 실제 응답 상태 확인
for i in {1..10}; do
  curl -w "\nHTTP Status: %{http_code}\n" http://rollouts-demo-bg-preview.example.com
  sleep 1
done

# 에러 로그 확인
kubectl logs -l app=rollouts-demo-bg -n default --tail=100
```

### 4. 롤백이 필요한 경우

```bash
# Rollout 이전 버전으로 롤백
kubectl argo rollouts undo rollouts-demo-bg -n default

# 또는 특정 revision으로 롤백
kubectl argo rollouts undo rollouts-demo-bg --to-revision=<revision-number> -n default
```

---

## 자동 승인 설정 (선택사항)

자동 승인을 원하는 경우 `blue-green.yaml`을 수정:

```yaml
autoPromotionEnabled: true
```

이 경우 Analysis가 성공하면 자동으로 Active로 전환됩니다.

---

## 참고사항

1. **트래픽 생성**: 실제 테스트에서는 더 많은 트래픽을 생성해야 Analysis가 정확하게 작동합니다.
2. **시간**: 각 Analysis 단계는 3분 동안 진행됩니다 (30초 간격으로 6번 확인).
3. **메트릭 수집**: Prometheus가 nginx ingress 메트릭을 수집하는데 시간이 걸릴 수 있습니다.
4. **DNS**: 로컬 테스트 시 `/etc/hosts`에 ingress IP를 추가해야 할 수 있습니다.

