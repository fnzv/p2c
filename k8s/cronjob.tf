resource "kubernetes_cron_job_v1" "prometheus2csv" {
  metadata {
    name = "prometheus2csv"
  }
  spec {
    concurrency_policy            = "Forbid"
    failed_jobs_history_limit     = 2
    schedule                      = "0 0 * * *"
    starting_deadline_seconds     = 10
    successful_jobs_history_limit = 2
    job_template {
      metadata {}
      spec {
        backoff_limit              = 2
        ttl_seconds_after_finished = 10
        template {
          metadata {}
          spec {
            container {
              name    = "prometheus2csv"
              image   = "fnzv/p2c"
              command = ["/usr/local/bin/p2c"] 
              env {
                  name = "P2C_ADDRESS"
                  value = "http://prometheus.ingress"
                }
              env {
                  name = "P2C_REGION"
                  value = "eu-west-1"
                }
              env {
                  name = "P2C_TIMERANGE"
                  value = "29d"
                }
              env {
                  name = "P2C_QUERY"
                  value = "combustionFuelLevel_range{job='bmw_exporter'} currentMileages{job='bmw_exporter'} gas_fuel_price{job='bmw_exporter'} combustionFuelLevel_remainingFuelLiters{job='bmw_exporter'} checkControlMessages_severity{job='bmw_exporter'} checkControlMessages_severity {job='bmw_exporter'}"
                }
              env {
                  name = "P2C_UPLOAD_S3"
                  value = "s3://your-s3-bucket/metrics/"
                }
              env {
                  name = "AWS_SECRET_ACCESS_KEY"
                  value = var.s3_access_key
                }
              env {
                  name = "AWS_ACCESS_KEY_ID"
                  value = var.s3_secret_key
                }
             
            }
          }
        }
      }
    }
  }
}
