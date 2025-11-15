# ArgoCD 초기 설정 아키텍처 문서

이 문서는 `argocd-init` 폴더의 구조와 cert-manager, ingress-nginx, ArgoCD가 서로 어떻게 연관되어 작동하는지 설명합니다.

## 목차
1. [전체 구조 개요](#전체-구조-개요)
2. [컴포넌트별 역할](#컴포넌트별-역할)
3. [설치 순서 및 의존성](#설치-순서-및-의존성)
4. [작동 흐름](#작동-흐름)
5. [파일 구조 설명](#파일-구조-설명)

## 전체 구조 개요

이 프로젝트는 **Bootstrap 패턴**을 사용하여 ArgoCD를 자체 관리(Self-Managed) 방식으로 설정합니다.

```
argocd-init/
├── bootstrap/          # 초기 설치를 위한 리소스 (수동 적용)
│   ├── app-of-apps.yaml
│   ├── cert-manager/
│   ├── ingress-nginx/
│   └── argo-cd/
└── apps/              # ArgoCD가 관리할 Application 정의
    ├── cert-manager/
    ├── ingress-nginx/
    ├── argocd/
    └── projects/
```

### Bootstrap vs Apps

- **Bootstrap**: ArgoCD가 설치되기 전에 수동으로 적용해야 하는 리소스들
- **Apps**: ArgoCD가 설치된 후, ArgoCD Application을 통해 자동으로 관리되는 리소스들

## 컴포넌트별 역할

### 1. cert-manager
**역할**: Kubernetes 클러스터에서 TLS 인증서를 자동으로 발급하고 관리

**주요 리소스**:
- `ClusterIssuer` (selfsigned-issuer): 자체 서명 인증서를 발급하는 Issuer
- ArgoCD의 Ingress에서 사용할 TLS 인증서를 자동으로 생성

**설치 위치**: `bootstrap/cert-manager/`

### 2. ingress-nginx
**역할**: Kubernetes Ingress Controller로 외부에서 클러스터 내부 서비스에 접근할 수 있게 함

**주요 기능**:
- 외부 트래픽을 클러스터 내부 서비스로 라우팅
- SSL/TLS 종료 처리
- ArgoCD 서버에 대한 외부 접근 제공

**설치 위치**: `bootstrap/ingress-nginx/`

### 3. ArgoCD
**역할**: GitOps를 통한 애플리케이션 배포 및 관리 플랫폼

**주요 기능**:
- Git 저장소의 변경사항을 감지하고 자동 동기화
- Kubernetes 리소스의 상태 모니터링
- 롤백 및 헬스 체크 기능

**설치 위치**: `bootstrap/argo-cd/`

## 설치 순서 및 의존성

### Phase 1: Bootstrap (수동 설치)

ArgoCD가 아직 설치되지 않은 상태이므로, 다음 순서로 수동으로 설치해야 합니다:

```
1. cert-manager 설치
   ↓
2. ingress-nginx 설치
   ↓
3. ArgoCD 설치 (cert-manager와 ingress-nginx에 의존)
   ↓
4. app-of-apps Application 생성
```

#### 1단계: cert-manager 설치
```bash
kubectl apply -k bootstrap/cert-manager
```

**설치 내용**:
- cert-manager v1.19.1 설치
- `selfsigned-issuer` ClusterIssuer 생성 (자체 서명 인증서 발급용)

#### 2단계: ingress-nginx 설치
```bash
kubectl apply -k bootstrap/ingress-nginx
```

**설치 내용**:
- ingress-nginx Controller 설치
- SSL passthrough 활성화
- Kind 환경을 위한 설정 적용

#### 3단계: ArgoCD 설치
```bash
kubectl apply -k bootstrap/argo-cd
```

**설치 내용**:
- ArgoCD 공식 설치 매니페스트 적용
- Ingress 리소스 생성 (argocd.example.com)
- Certificate 리소스 생성 (cert-manager가 TLS 인증서 발급)
- ArgoCD 서버 설정 (insecure 모드, basehref 설정)
- Rollout extension 설치

**의존성**:
- cert-manager: Certificate 리소스가 TLS 인증서를 자동 생성하기 위해 필요
- ingress-nginx: Ingress 리소스가 작동하기 위해 필요

#### 4단계: app-of-apps 패턴 적용
```bash
kubectl apply -f bootstrap/app-of-apps.yaml
```

이 Application은 `apps/` 디렉토리의 모든 Application을 자동으로 생성합니다.

### Phase 2: Apps (ArgoCD 자동 관리)

ArgoCD가 설치된 후, `apps/` 디렉토리의 Application들이 자동으로 생성되어 각 컴포넌트를 관리합니다:

```
app-of-apps (bootstrap/app-of-apps.yaml)
    ├── cert-manager Application
    ├── ingress-nginx Application
    └── argocd Application
```

각 Application은 해당하는 `bootstrap/` 디렉토리의 리소스를 GitOps 방식으로 관리합니다.

## 작동 흐름

### 1. TLS 인증서 발급 흐름

```
ArgoCD Ingress 생성
    ↓
Certificate 리소스 생성 (bootstrap/argo-cd/certificate.yaml)
    ↓
cert-manager가 Certificate 리소스 감지
    ↓
ClusterIssuer (selfsigned-issuer)를 통해 인증서 발급
    ↓
argocd-tls Secret 생성 (인증서 저장)
    ↓
Ingress가 Secret을 참조하여 HTTPS 트래픽 처리
```

**관련 파일**:
- `bootstrap/argo-cd/certificate.yaml`: Certificate 리소스 정의
- `bootstrap/cert-manager/selfsigned-issuer.yaml`: ClusterIssuer 정의
- `bootstrap/argo-cd/ingress.yaml`: Ingress에서 TLS Secret 참조

### 2. 외부 접근 흐름

```
사용자 요청 (https://argocd.example.com)
    ↓
ingress-nginx Controller가 요청 수신
    ↓
Ingress 규칙에 따라 argocd-server Service로 라우팅
    ↓
ArgoCD 서버가 요청 처리
```

**관련 파일**:
- `bootstrap/ingress-nginx/kustomization.yaml`: ingress-nginx Controller 설치
- `bootstrap/argo-cd/ingress.yaml`: ArgoCD Ingress 규칙 정의

### 3. GitOps 관리 흐름

```
Git 저장소 변경
    ↓
app-of-apps Application이 변경 감지
    ↓
apps/ 디렉토리의 Application들 생성/업데이트
    ↓
각 Application이 bootstrap/ 디렉토리의 리소스 동기화
    ↓
Kubernetes 클러스터에 리소스 적용
```

**관련 파일**:
- `bootstrap/app-of-apps.yaml`: App-of-Apps 패턴의 루트 Application
- `apps/*/application.yaml`: 각 컴포넌트의 Application 정의

## 파일 구조 설명

### Bootstrap 디렉토리

#### `bootstrap/cert-manager/`
- **kustomization.yaml**: cert-manager v1.19.1 설치 및 ClusterRole 패치
- **selfsigned-issuer.yaml**: 자체 서명 인증서를 발급하는 ClusterIssuer

#### `bootstrap/ingress-nginx/`
- **kustomization.yaml**: ingress-nginx Controller 설치 및 Kind 환경 설정

#### `bootstrap/argo-cd/`
- **namespace.yaml**: argocd 네임스페이스 생성
- **kustomization.yaml**: ArgoCD 설치 및 서버 설정 패치
- **ingress.yaml**: ArgoCD 서버에 대한 Ingress 규칙 (TLS 포함)
- **certificate.yaml**: cert-manager를 통한 TLS 인증서 발급 요청

#### `bootstrap/app-of-apps.yaml`
- ArgoCD Application의 Application (App-of-Apps 패턴)
- `apps/` 디렉토리의 모든 Application을 자동 생성

### Apps 디렉토리

#### `apps/cert-manager/application.yaml`
- cert-manager를 관리하는 ArgoCD Application
- `bootstrap/cert-manager/` 경로를 소스로 사용
- 자동 동기화 및 자가 치유 활성화

#### `apps/ingress-nginx/application.yaml`
- ingress-nginx를 관리하는 ArgoCD Application
- `bootstrap/ingress-nginx/` 경로를 소스로 사용
- Admission Job은 무시하도록 설정 (일회성 리소스)

#### `apps/argocd/application.yaml`
- ArgoCD 자체를 관리하는 ArgoCD Application (Self-Managed)
- `bootstrap/argo-cd/` 경로를 소스로 사용
- 자동 동기화 및 자가 치유 활성화

#### `apps/projects/infra.yaml`
- ArgoCD AppProject 정의
- 인프라 컴포넌트 관리를 위한 프로젝트 설정

## 주요 설정 상세

### ArgoCD Ingress 설정

```yaml
# bootstrap/argo-cd/ingress.yaml
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - argocd.example.com
    secretName: argocd-tls  # cert-manager가 생성한 Secret
```

### Certificate 설정

```yaml
# bootstrap/argo-cd/certificate.yaml
spec:
  secretName: argocd-tls
  issuerRef:
    name: selfsigned-issuer
    kind: ClusterIssuer
  dnsNames:
  - argocd.example.com
```

### ArgoCD 서버 설정

```yaml
# bootstrap/argo-cd/kustomization.yaml의 패치
data:
  server.insecure: "true"  # Ingress를 통한 접근 허용
  server.basehref: "/"     # 경로 설정
```

## 의존성 다이어그램

```
┌─────────────────┐
│  cert-manager   │
│  (Bootstrap)    │
└────────┬────────┘
         │
         │ TLS 인증서 발급
         │
         ▼
┌─────────────────┐
│  ingress-nginx  │
│  (Bootstrap)    │
└────────┬────────┘
         │
         │ 외부 접근 제공
         │
         ▼
┌─────────────────┐
│     ArgoCD      │
│  (Bootstrap)    │
│                 │
│  ┌───────────┐  │
│  │ Ingress   │──┼──► cert-manager (TLS)
│  └───────────┘  │
└────────┬────────┘
         │
         │ App-of-Apps 패턴
         │
         ▼
┌─────────────────┐
│  app-of-apps    │
│  Application    │
└────────┬────────┘
         │
         ├──► cert-manager Application
         ├──► ingress-nginx Application
         └──► argocd Application (Self-Managed)
```

## 주의사항

1. **설치 순서**: cert-manager → ingress-nginx → ArgoCD 순서로 설치해야 합니다.
2. **DNS 설정**: `argocd.example.com`을 실제 도메인으로 변경하거나 `/etc/hosts`에 추가해야 합니다.
3. **Self-Managed**: ArgoCD가 자신을 관리하므로, bootstrap 단계에서만 수동 적용이 필요합니다.
4. **인증서 유형**: 현재는 self-signed 인증서를 사용합니다. 프로덕션 환경에서는 Let's Encrypt 등 실제 CA를 사용하는 것이 좋습니다.

## 참고사항

- 이 구조는 **App-of-Apps 패턴**을 사용하여 여러 Application을 효율적으로 관리합니다.
- ArgoCD는 **Self-Managed** 방식으로 자신을 관리하므로, 초기 bootstrap 후에는 GitOps를 통해 자동으로 관리됩니다.
- cert-manager와 ingress-nginx는 ArgoCD가 설치되기 전에 필요한 인프라 컴포넌트이므로 bootstrap 단계에서 설치됩니다.

