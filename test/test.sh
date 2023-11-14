#! /bin/sh
# 检查kubectl和helm版本，kubectl client和server版本都要检查
minor_version=$(kubectl version --short | grep -oE '[0-9]+\.[0-9]+' | tr '.' ' ' | cut -f2 -d ' ' )
if [ $? = 1 ] ; then
  echo "kubectl is not installed" ; exit 1
else
  for v in ${minor_version} ;
  do
    if ((${v}<14)) ; then
      echo "kubectl version must >=1.14!" ; exit 1
    fi
  done
fi

helm version
if [ $? = 1 ] ; then
  echo "helm is not installed" ; exit 1
fi

# 检查有没有运行中的fluid任务,如果有的话提示并退出(多个fluid会引起端口冲突),没有的话创建fluid环境
# 可以通过helm delete fluid 来结束运行中的fluid任务

if [ -n "$(helm list --all-namespaces -a | grep fluid )" ] ; then
  echo "before running test,you have to make sure there is no running fluid" ; exit 1
fi

helm install fluid ../charts/fluid/fluid

# 场景1 测试dataset基本的数据访问功能,job中的容器会挂载dataset,并访问其中的数据
# 预期是test_job成功执行
kubectl create -f testcase1/


for ((i=0;i<200;i++));
do
  sleep 1
  status=$(kubectl get job | grep fluid-test |  grep -o '[0-9]\/[0-9]')
  if [ ${status} = '1/1' ] ; then
    success="1"
    break
  fi
done

if [ ${success} != '1' ];then
  echo "test1 failed !" ; exit 1
fi

echo "first test passed!!"
kubectl delete -f testcase1/

# 场景2 测试亲和性feature

# testcase2_dataset.yaml设置为只绑定到fluid-test为true的node上
# 启动dataset,不为节点设置label,dataset预期状态为not bound
kubectl create -f testcase2/
sleep 5
status=$( kubectl get dataset ant  | tail -1 | awk -F ' ' '{print$2}' )

if [ "${status}"  != "NotBound" ] ;then
  echo ${status};echo "phase should be 'not bound'!"; exit 1
fi

node=$( kubectl get node | grep 'Ready' | cut -f1 -d ' ' | head -1 )
kubectl label node "${node}" fluid-test=true
# dataset绑定runtime需要一些时间..
sleep 100

# 为集群中一个任意一个Ready节点打label,dataset预期状态为Bound
status=$(kubectl get dataset ant  | tail -1 | awk -F ' ' '{print$6}')
if [ "${status}" !=  "Bound" ] ;then
  echo ${status};echo "phase should be bound!"; exit 1
fi

echo "test2 passed!"
kubectl delete -f testcase2/
helm delete fluid
kubectl label node ${node} fluid-test-

# 目前异常退出时,没有释放fluid资源,需要fix
