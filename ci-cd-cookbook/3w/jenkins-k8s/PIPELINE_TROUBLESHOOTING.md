# Jenkins Pipeline 무한 로딩 문제 해결 가이드

## 문제 원인

Jenkins Pipeline에서 무한 로딩이 발생하는 주요 원인은 다음과 같습니다:

1. **네임스페이스 불일치**: ServiceAccount `jenkins`는 `jenkins` 네임스페이스에 있지만, Pod 스펙에서 네임스페이스가 명시되지 않음
2. **Pod 생성 실패**: Kubernetes agent가 Pod를 생성하지 못함
3. **JNLP 연결 실패**: Pod가 생성되어도 Jenkins controller와 연결하지 못함

## 해결 방법

### 1. Pipeline 수정사항

주요 변경사항:
- `namespace 'jenkins'` 추가: Kubernetes agent가 Pod를 생성할 네임스페이스 명시
- Pod 스펙에 `metadata.namespace: jenkins` 추가
- 리소스 제한 추가: Pod가 제대로 스케줄되도록 리소스 요청/제한 추가

### 2. Jenkins Kubernetes 플러그인 설정 확인

Jenkins 관리 화면에서 확인:
1. **Manage Jenkins** → **Configure System** → **Cloud** → **Kubernetes**
2. 다음 설정 확인:
   - **Name**: kubernetes
   - **Kubernetes URL**: `https://kubernetes.default.svc` (클러스터 내부)
   - **Kubernetes Namespace**: `jenkins`
   - **Credentials**: ServiceAccount 토큰 자동 사용 또는 명시적 설정
   - **Jenkins URL**: Jenkins controller의 내부 URL
   - **Jenkins tunnel**: JNLP 연결을 위한 tunnel 설정 (필요시)

### 3. ServiceAccount 및 RBAC 확인

다음 명령으로 확인:
```bash
# ServiceAccount 존재 확인
kubectl get serviceaccount jenkins -n jenkins

# ClusterRoleBinding 확인
kubectl get clusterrolebinding jenkins

# ServiceAccount가 올바른 권한을 가지고 있는지 확인
kubectl auth can-i create pods --as=system:serviceaccount:jenkins:jenkins -n jenkins
kubectl auth can-i list pods --as=system:serviceaccount:jenkins:jenkins -n jenkins
```

### 4. Pod 생성 및 JNLP 연결 확인

Pipeline 실행 시 다음을 확인:
```bash
# jenkins 네임스페이스에서 생성되는 Pod 확인
kubectl get pods -n jenkins

# Pod 로그 확인 (JNLP 에이전트)
kubectl logs <pod-name> -n jenkins -c jnlp

# Jenkins controller 로그 확인
kubectl logs <jenkins-controller-pod> -n jenkins
```

## 추가 체크리스트

### Kubernetes 클러스터 연결
- [ ] Jenkins controller가 Kubernetes API 서버에 접근 가능한지 확인
- [ ] Kubernetes 서비스 계정이 올바르게 설정되었는지 확인

### 네트워크 설정
- [ ] Jenkins controller의 JNLP 포트(50000)가 열려있는지 확인
- [ ] 클러스터 내부 DNS가 정상 작동하는지 확인

### 이미지 접근
- [ ] `jenkins/inbound-agent:latest` 이미지가 클러스터에서 접근 가능한지 확인
- [ ] `bitnami/kubectl:latest` 이미지가 클러스터에서 접근 가능한지 확인

### 리소스 제한
- [ ] 클러스터에 충분한 리소스(CPU, Memory)가 있는지 확인
- [ ] Node Selector/Affinity 문제가 없는지 확인

## 디버깅 명령어

```bash
# Pod 이벤트 확인
kubectl describe pod <pod-name> -n jenkins

# ServiceAccount 상세 확인
kubectl describe serviceaccount jenkins -n jenkins

# ClusterRole 상세 확인
kubectl describe clusterrole jenkins

# Jenkins controller 로그 확인
kubectl logs -f <jenkins-controller-pod> -n jenkins

# Kubernetes API 서버 연결 테스트
kubectl cluster-info
```

## 예상되는 오류 메시지

1. **"Failed to create pod"**: ServiceAccount 권한 부족 또는 네임스페이스 접근 불가
2. **"Connection refused"**: JNLP 포트 미접근 또는 Jenkins URL 설정 오류
3. **"ImagePullBackOff"**: 이미지 다운로드 실패
4. **"Pending"**: 리소스 부족 또는 Node Selector 문제
