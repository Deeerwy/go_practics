# Практическое занятие 16 — Публикация сервиса tasks в Kubernetes. Бурылин Дмитрий ПИМО-01-25

## Цель работы

Освоить базовую публикацию контейнеризированного backend-приложения в Kubernetes:
описать Deployment и Service, передать конфигурацию через ConfigMap,
настроить readiness и liveness probes, применить манифесты через kubectl
и проверить работу сервиса через port-forward.

## Структура проекта

```
pz16-kuber/
├── main.go
├── go.mod
├── Dockerfile
└── deploy/
    └── k8s/
        ├── configmap.yaml
        ├── deployment.yaml
        └── service.yaml
```

## Описание сервиса

Сервис tasks — минимальный HTTP-сервер на Go.

Эндпоинты:

- GET /health — проверка состояния сервиса (используется для readiness и liveness probes)

Конфигурация передаётся через переменные окружения из ConfigMap:

- TASKS_PORT    — порт, на котором слушает сервис (по умолчанию 8082)
- AUTH_BASE_URL — адрес сервиса авторизации
- LOG_LEVEL     — уровень логирования


## Подготовка образа для Kubernetes

Образ собирается внутри Docker-демона minikube,
чтобы кластер мог получить его без внешнего registry.

Шаг 1 — Переключить терминал на Docker-демон minikube:

```bash
eval $(minikube docker-env)
```

Шаг 2 — Собрать образ с фиксированным тегом:

```bash
docker build -t techip-tasks:0.1 .
```

Шаг 3 — Убедиться, что образ доступен:

```bash
docker images | grep techip-tasks
```

Способ обеспечения доступности образа кластеру:
пересборка образа внутри Docker-демона minikube через eval $(minikube docker-env).
В Deployment указан imagePullPolicy: IfNotPresent,
поэтому Kubernetes берёт образ из локального кэша и не обращается к внешнему registry.


## Манифесты Kubernetes

### ConfigMap — deploy/k8s/configmap.yaml

Хранит несекретную конфигурацию приложения.
Все переменные передаются в контейнер через envFrom.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tasks-config
data:
  TASKS_PORT: "8082"
  AUTH_BASE_URL: "http://auth:8081"
  LOG_LEVEL: "info"
```

### Deployment — deploy/k8s/deployment.yaml

Описывает желаемое состояние: одна реплика сервиса tasks.
При сбое Pod Kubernetes автоматически создаёт новый.
Настроены readiness и liveness probes через GET /health.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tasks
  labels:
    app: tasks
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tasks
  template:
    metadata:
      labels:
        app: tasks
    spec:
      containers:
        - name: tasks
          image: techip-tasks:0.1
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8082
          envFrom:
            - configMapRef:
                name: tasks-config
          readinessProbe:
            httpGet:
              path: /health
              port: 8082
            initialDelaySeconds: 3
            periodSeconds: 5
          livenessProbe:
            httpGet:
              path: /health
              port: 8082
            initialDelaySeconds: 10
            periodSeconds: 10
          resources:
            requests:
              memory: "32Mi"
              cpu: "50m"
            limits:
              memory: "64Mi"
              cpu: "100m"
```

### Service — deploy/k8s/service.yaml

Обеспечивает стабильную точку доступа к Pod внутри кластера.
Тип ClusterIP — доступ только изнутри кластера.
Для демонстрации извне используется kubectl port-forward.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: tasks
  labels:
    app: tasks
spec:
  type: ClusterIP
  selector:
    app: tasks
  ports:
    - protocol: TCP
      port: 8082
      targetPort: 8082
```


## Применение манифестов

```bash
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
```
![1](./screens/image.png)

## Проверка состояния ресурсов

Проверка Pod:

```bash
kubectl get pods
kubectl describe pod tasks-7d6cdd64bc-zwkqh
```
![2](./screens/image-1.png)

Проверка Deployment:

```bash
kubectl get deployment
kubectl describe deployment tasks
```
![3](./screens/image-2.png)

Проверка Service:

```bash
kubectl get svc
kubectl describe svc tasks
```
![4](./screens/image-3.png)

Логи контейнера:

```bash
kubectl logs tasks-67d65c5bfd-w4mls
```
![5](./screens/image-4.png)


## Демонстрация доступа через port-forward

Терминал 1 — пробросить порт:

```bash
kubectl port-forward svc/tasks 8082:8082
```

Терминал 2 — проверить эндпоинты:

```bash
curl -i http://localhost:8082/health

```
![6](./screens/image-5.png)

## Масштабирование

Увеличить число реплик до 2:

```bash
kubectl scale deployment tasks --replicas=2
kubectl get pods
```
![7](./screens/image-6.png)

Вернуть к одной реплике:

```bash
kubectl scale deployment tasks --replicas=1
kubectl get pods
```
![8](./screens/image-7.png)

## Удаление ресурсов

```bash
kubectl delete -f deploy/k8s/service.yaml
kubectl delete -f deploy/k8s/deployment.yaml
kubectl delete -f deploy/k8s/configmap.yaml
```


## Ответы на контрольные вопросы

### 1. Что такое Kubernetes и для чего он используется?

Kubernetes — система оркестрации контейнеров. Она автоматически запускает,
масштабирует и сопровождает приложения в контейнерах: поддерживает нужное число
экземпляров, перезапускает упавшие Pod, отделяет конфигурацию от образа
и предоставляет единый способ публикации сервисов.

### 2. Чем Pod отличается от Deployment?

Pod — минимальная единица запуска в Kubernetes, один или несколько контейнеров,
запущенных вместе. Deployment — управляющий объект более высокого уровня:
он описывает, сколько Pod должно работать и как, и следит за тем, чтобы
фактическое состояние соответствовало желаемому.

### 3. Почему приложение публикуют через Deployment, а не через одиночный Pod?

Одиночный Pod не восстанавливается автоматически при сбое — его нужно
пересоздавать вручную. Deployment отслеживает число живых реплик и при падении
Pod сразу создаёт новый, обеспечивая непрерывную работу сервиса.

### 4. Зачем нужен Service и почему нельзя обращаться к Pod напрямую?

IP-адрес и имя Pod нестабильны: при пересоздании Pod получает новый адрес.
Service предоставляет постоянную точку входа и автоматически направляет трафик
к актуальным Pod через selector по метке, независимо от их жизненного цикла.

### 5. Что такое ConfigMap?

ConfigMap — объект Kubernetes для хранения несекретной конфигурации приложения
в виде пар ключ-значение. Позволяет передавать параметры (порт, адреса сервисов,
уровень логирования) в контейнер через переменные окружения, не встраивая их
в Docker-образ.

### 6. Чем ConfigMap отличается от Secret?

ConfigMap хранит данные в открытом виде и предназначен для несекретной
конфигурации. Secret хранит чувствительные данные (пароли, токены, ключи)
в закодированном виде (base64) с ограниченным доступом. Использовать ConfigMap
для паролей — неправильно с точки зрения безопасности.

### 7. Для чего используется readiness probe?

Readiness probe проверяет, готово ли приложение принимать трафик. Пока проверка
не проходит успешно, Kubernetes не направляет на Pod запросы через Service.
Это важно при медленном старте: приложение не получит трафик раньше,
чем будет готово его обработать.

### 8. Для чего используется liveness probe?

Liveness probe проверяет, что приложение не зависло и продолжает корректно
работать. Если проверка стабильно проваливается, Kubernetes считает контейнер
неработоспособным и перезапускает его, восстанавливая работоспособность сервиса.

### 9. Почему важно использовать фиксированный тег образа, а не latest?

Тег latest не позволяет определить, какая именно версия приложения запущена
в кластере. Фиксированный тег (например, 0.1 или хэш коммита) обеспечивает
воспроизводимость развёртывания, упрощает поиск ошибок и делает обновления
прозрачными и управляемыми.

### 10. Зачем нужен kubectl port-forward?

kubectl port-forward создаёт туннель между локальным портом на машине разработчика
и портом Service или Pod внутри кластера. Это позволяет обращаться к сервису
с типом ClusterIP локально без настройки LoadBalancer или Ingress.

### 11. Что делает команда kubectl scale deployment?

Команда изменяет число реплик Deployment в реальном времени. Kubernetes
автоматически создаёт или удаляет Pod, чтобы привести фактическое число
экземпляров к указанному значению. Изменение происходит без пересоздания
самого объекта Deployment.

### 12. Почему публикация приложения в Kubernetes считается декларативной?

При декларативном подходе разработчик описывает желаемое состояние системы
в YAML-манифестах (сколько реплик, какой образ, какие probes), а не пишет
пошаговые инструкции по достижению этого состояния. Kubernetes сам сравнивает
текущее состояние кластера с желаемым и предпринимает необходимые действия
для их совпадения.
