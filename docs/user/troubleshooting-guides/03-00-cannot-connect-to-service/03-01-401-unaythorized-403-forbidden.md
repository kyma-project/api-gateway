# 401 Unauthorized or 403 Forbidden

## Symptom

When you try to reach your service, you get `401 Unauthorized` or `403 Forbidden` in response.

## Cause 

The error `401 Unauthorized` occurs when you try to access a Service that requires authentication, but you have either not provided appropriate credentials or have not provided any credentials at all. You get the error `403 Forbidden` when you try to access a Service or perform an action for which you lack permission.

## Remedy

Make sure that you are using an access token with proper scopes, and it is active. Depending on the type of your access token, follow the relevant steps.

### JWT token

TBD

### Opaque token

TBD