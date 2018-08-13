# Deploying to cluster
  ```
  1. kubectl apply -f example/operator.yaml
  2. kubectl apply -f example/cr.yaml
  ```
  
  ### Note: 
  1. If the namespace in which CR was deployed in is deleted, the s3 bucket gets deleted.
  2. If the CR from the namespace is deleted, the s3 bucket gets deleted
----

## TODO

##### 1. S3 sync
##### 2. Use labels to tag buckets ( add namespace )
##### 3. Allow more configuration options but default to something ( storageType, versioned, Bucket ACL, etc )
