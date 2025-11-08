# Blue/Green 배포 데모

이 디렉토리는 Argo Rollouts를 사용한 블루/그린 배포 실습 예제입니다.

## 개요

블루/그린 배포는 신규 버전을 기존 버전과 동일한 환경에 구성하여 테스트 이후 신규 버전으로 전환하는 배포 방식입니다.

**장점:**
- 테스트 후 버전 변경이 가능
- 빠른 롤백이 가능
- 무중단 배포

## 구조

```
blue-green/
├── rollout.yaml          # Argo Rollout 리소스 정의
├── active-service.yaml   # 현재 운영 중인 버전에 연결되는 서비스
├── preview-service.yaml  # 새로운 버전에 연결되는 서비스
└── kustomization.yaml    # Kustomize 설정
```

## 실습 가이드

### 1. 사전 준비

Argo Rollouts가 설치되어 있어야 합니다. 설치되어 있지 않다면:

```bash
kubectl create namespace argo-rollouts
kubectl apply -n argo-rollouts -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml
```

Argo Rollouts CLI를 설치합니다:

```bash
# macOS
brew install argoproj/tap/kubectl-argo-rollouts

# Linux
curl -LO https://github.com/argoproj/argo-rollouts/releases/latest/download/kubectl-argo-rollouts-linux-amd64
chmod +x ./kubectl-argo-rollouts-linux-amd64
sudo mv ./kubectl-argo-rollouts-linux-amd64 /usr/local/bin/kubectl-argo-rollouts
```

### 2. 배포 확인

Argo CD를 통해 배포되면, 다음 명령어로 상태를 확인할 수 있습니다:

```bash
# Rollout 상태 확인
kubectl argo rollouts get rollout blue-green-demo

# 상세 정보 확인
kubectl argo rollouts describe rollout blue-green-demo
```

### 3. 새 버전 배포 (Green 배포)

새 버전을 배포하려면 `rollout.yaml`의 이미지를 변경합니다:

```yaml
spec:
  template:
    spec:
      containers:
      - name: blue-green-demo
        image: argoproj/rollouts-demo:green  # blue -> green으로 변경
        env:
        - name: DEMO_COLOR
          value: "green"
```

변경사항을 Git에 커밋하고 푸시하면 Argo CD가 자동으로 감지하여 새 버전을 배포합니다.

### 4. Preview Service로 테스트

새 버전이 배포되면 Preview Service를 통해 테스트할 수 있습니다:

```bash
# Preview Service로 포트 포워딩
kubectl port-forward svc/blue-green-demo-preview 8080:80

# 다른 터미널에서 테스트
curl http://localhost:8080
```

### 5. 프로모션 (Green으로 전환)

테스트가 완료되면 수동으로 프로모션을 진행합니다:

```bash
# 프로모션 실행
kubectl argo rollouts promote blue-green-demo

# 또는 Argo CD UI에서 Rollout 리소스를 확인하고 프로모션 버튼 클릭
```

### 6. 롤백

문제가 발생하면 빠르게 롤백할 수 있습니다:

```bash
# 롤백 실행
kubectl argo rollouts undo blue-green-demo

# 특정 리비전으로 롤백
kubectl argo rollouts undo blue-green-demo --to-revision=2
```

## 주요 설정 설명

### Rollout 설정

- `autoPromotionEnabled: false`: 자동 프로모션 비활성화 (수동 승인 필요)
- `scaleDownDelaySeconds: 30`: 이전 버전 파드 종료 전 대기 시간
- `antiAffinity`: 파드 분산 배치를 위한 설정

### Service 설정

- `active-service`: 현재 운영 중인 버전 (Blue)에 트래픽을 라우팅
- `preview-service`: 새 버전 (Green)에 트래픽을 라우팅 (테스트용)

## 실습 시나리오

1. **초기 배포**: Blue 버전이 Active Service에 연결되어 배포됨
2. **새 버전 배포**: Green 버전이 Preview Service에 연결되어 배포됨
3. **테스트**: Preview Service를 통해 Green 버전 테스트
4. **프로모션**: 테스트 완료 후 Green 버전을 Active로 전환
5. **정리**: Blue 버전 파드가 자동으로 종료됨

## 참고 자료

- [Argo Rollouts 공식 문서](https://argoproj.github.io/argo-rollouts/)
- [블루/그린 배포 블로그](https://devocean.sk.com/blog/techBoardDetail.do?ID=165522&boardType=techBlog)
- [Argo Rollouts 예제](https://github.com/argoproj/rollouts-demo)

