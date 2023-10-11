# Ory limitations

## Resource configuration

By default, the Ory components' resources have the following configuration for bigger clusters:

| Component          |          | CPU         | Memory |
|--------------------|----------|-------------|--------|
| Oathkeeper         | Limits   | 10 (10000m) | 512Mi  |
| Oathkeeper         | Requests | 100m        | 64Mi   |
| Oathkeeper Maester | Limits   | 400m        | 1Gi    |
| Oathkeeper Maester | Requests | 10m         | 32Mi   |

For smaller clusters with less then 5 Cpu capability or less then 10 Gi of memory capability, the Ory components' resources have the following configuration:

| Component          |          | CPU  | Memory |
|--------------------|----------|------|--------|
| Oathkeeper         | Limits   | 100m | 128Mi  |
| Oathkeeper         | Requests | 10m  | 64Mi   |
| Oathkeeper Maester | Limits   | 100m | 50Mi   |
| Oathkeeper Maester | Requests | 10m  | 20Mi   |


## Autoscaling configuration

The default configuration in terms of autoscaling of Ory components is as follows:

| Component          | Min replicas       | Max replicas       |
|--------------------|--------------------|--------------------|
| Oathkeeper         | 3                  | 10                 |
| Oathkeeper Maester | Same as Oathkeeper | Same as Oathkeeper |

Oathkeeper Maester is set up as a separate container in the same Pod as Oathkeeper. Because of that, their autoscaling configuration is similar.