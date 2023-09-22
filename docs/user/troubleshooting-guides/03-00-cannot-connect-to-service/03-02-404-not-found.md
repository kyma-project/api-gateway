# 404 Not Found

## Symptom

When you try to reach your Service, you get `404 Not Found` in response.

## Remedy

Make sure that the following conditions are met:

- The proper Oathkeeper Acess Rule has been created:

  ```bash
  kubectl get rules.oathkeeper.ory.sh -n {NAMESPACE}
  ```

  >**TIP:** The name of the Access Rule consists of the name of the APIRule and a random suffix.

- Proper VirtualService has been created:

  ```bash
  kubectl get virtualservices.networking.istio.io -n {NAMESPACE}
  ```

  >**TIP:** The name of the VirtualService consists of the name of the APIRule and a random suffix.

Sometimes, the Oathkeeper Maester controller stops reconciling Access Rules on long-living clusters. This behavior might result in the random `404 Not Found` responses because Oathkeeper does not contain Access Rules reflecting the actual state of the cluster. A simple restart of the Pod resolves the problem. To verify if that it is the cause of the issue you encountered, follow these steps:

1. Fetch all Oathkeeper Pods' names:

    ```bash
    kubectl get pods -n kyma-system -l app.kubernetes.io/name=oathkeeper -o jsonpath='{.items[*].metadata.name}'
    ```

2. Fetch the Access Rules from every Oathkeeper Pod and save them to a file:

    ```bash
   kubectl cp -n kyma-system -c oathkeeper "{POD_NAME}":etc/rules/access-rules.json "access-rules.{POD_NAME}.json" 
   ```

3. If you have more than one instance of Oathkeeper, compare whether the files contain the same Access Rules. Because Oathkeeper stores Access Rules as JSON files, you can use [jd](https://github.com/josephburnett/jd) to automate the comparison:

    ```bash
   jd -set {FIRST_FILE} {SECOND_FILE} 
   ```

    The files are considered different by jd if there are any differences between the files other than the order of Access Rules.
   
4. Compare Access Rules in the files with those present in the cluster. If the files are different, Oathkeeper Pods are out of sync.
