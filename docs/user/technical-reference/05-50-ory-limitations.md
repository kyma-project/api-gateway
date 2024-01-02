# Ory Limitations

## Resource Configuration
The resources of the Ory components are configured differently based on the size of the cluster. Smaller clusters are defined as having less than 5 CPU capability or less than 10 Gi of memory capability, while larger clusters exceed these values.

The default configuration for larger clusters includes the following settings for the Ory components' resources:

| Component          |          | CPU         | Memory |
|--------------------|----------|-------------|--------|
| Oathkeeper         | Limits   | 10 (10000m) | 512Mi  |
| Oathkeeper         | Requests | 100m        | 64Mi   |
| Oathkeeper Maester | Limits   | 400m        | 1Gi    |
| Oathkeeper Maester | Requests | 10m         | 32Mi   |

The default configuration for smaller clusters includes the following settings for the Ory components' resources:

| Component          |          | CPU  | Memory |
|--------------------|----------|------|--------|
| Oathkeeper         | Limits   | 100m | 128Mi  |
| Oathkeeper         | Requests | 10m  | 64Mi   |
| Oathkeeper Maester | Limits   | 100m | 50Mi   |
| Oathkeeper Maester | Requests | 10m  | 20Mi   |


## Autoscaling Configuration

The default configuration in terms of autoscaling of Ory components is as follows:

| Component          | Min replicas       | Max replicas       |
|--------------------|--------------------|--------------------|
| Oathkeeper         | 3                  | 10                 |
| Oathkeeper Maester | Same as Oathkeeper | Same as Oathkeeper |

Oathkeeper Maester is set up as a separate container in the same Pod as Oathkeeper. Because of that, their autoscaling configuration is similar.