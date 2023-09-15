# Create a workload

This tutorial explains how to create a sample HttpBin Service deployment and a sample Function.

## Prerequisites
- [Install the Serverless module](https://github.com/kyma-project/serverless-manager/blob/main/docs/contributor/01-10-manual-installation-on-k3d.md) to be able to deploy a Function

## Steps

1. Create a Namespace and export its value as an environment variable. Run:

   ```bash
   export NAMESPACE={NAMESPACE_NAME}
   kubectl create ns $NAMESPACE
   kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
   ```
2. Deploy an instance of the HttpBin Service or a sample Function.
   
  <div tabs name="create">

    <details>
    <summary>
    HttpBin
    </summary>

    To deploy an instance of the HttpBin Service in your Namespace using the [sample code](https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml), run:

    ```shell
    kubectl -n $NAMESPACE create -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
    ```

    </details>

    <details>
    <summary>
    Function
    </summary>

    To create a Function in your Namespace using the [sample code](https://raw.githubusercontent.com/kyma-project/kyma/main/docs/03-tutorials/00-api-exposure/assets/function.yaml), run:

    ```shell
    kubectl -n $NAMESPACE apply -f https://raw.githubusercontent.com/kyma-project/kyma/main/docs/03-tutorials/00-api-exposure/assets/function.yaml
    ```

    </details>
  </div>

3. Verify if an instance of the HttpBin Service or a sample Function is successfully created.
   
  <div tabs name="verify">

    <details>
    <summary>
    HttpBin
    </summary>

    To verify if an instance of the HttpBin Service is created, run:

      ```shell
        kubectl get pods -l app=httpbin -n $NAMESPACE
      ```
    
    You should get a result similar to this one:
    
      ```shell
        NAME             READY    STATUS     RESTARTS    AGE
        httpbin-test     2/2      Running    0           96s
      ```

    </details>

    <details>
    <summary>
    Function
    </summary>

    To verify if a Function is created, run:

      ```shell
        kubectl get functions $NAME -n $NAMESPACE
      ```

    You should get a result similar to this one:
    
      ```shell
        NAME            CONFIGURED   BUILT     RUNNING   RUNTIME    VERSION   AGE
        test-function   True         True      True      nodejs18   1         96s
      ```
    </details>
  </div>