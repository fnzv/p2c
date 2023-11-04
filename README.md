
# Prometheus 2 CSV Exporter

  ![Project Logo](/img/car-report-sheet.PNG)

This is a Go application that exports data from Prometheus as a CSV file and optionally uploads it to an AWS S3 bucket. This README will guide you through setting up and running the application.

Related blog post:
- https://blog.sami.pw/2023/11/exporting-prometheus-metrics-into.html
  

## Table of Contents


- [Prometheus 2 CSV Exporter](#prometheus-2-csv-exporter)
  - [Table of Contents](#table-of-contents)
  - [Prerequisites](#prerequisites)
  - [Demo](#demo)
  - [Installation](#installation)
  - [Build](#build)
  - [Examples](#examples)
  - [Contributing](#contributing)
  - [License](#license)

  



## Prerequisites

  

Before using this exporter, you should have the following prerequisites installed and configured:

- Go (tested with Golang 1.21.1): Make sure you have Go installed on your system,  you can download and install Go from the [official Go website](https://golang.org/doc/install).

- AWS S3 bucket (Optional): If you plan to upload the CSV file to an AWS S3 bucket, make sure you have the AWS credentials configured, the Go script will automatically load the default AWS credentials or take them from enviroment variables `AWS_SECRET_ACCESS_KEY` and `AWS_ACCESS_KEY_ID`

- Prometheus: You must have access to a Prometheus server with a valid address to query data.


## Demo

Tested on Ubuntu 22.04 and Golang 1.18, 1.21

[![asciicast](https://asciinema.org/a/dnnGv3ZYarAeXER37uYTKYdZt.png)](https://asciinema.org/a/dnnGv3ZYarAeXER37uYTKYdZt)



## Installation

 

1. Clone this repository to your local machine:

  
`git clone https://github.com/fnzv/p2c.git`

  

Change your working directory to the project folder:

  

`cd p2c/`

## Build

 
`go build p2c.go`
  

The exporter binary will be created in the project folder.

`./p2c`
  

You can run the Prometheus to CSV exporter using the following command:

  

  
```
./p2c --query <Prometheus  query> --time-range <time range> --address <Prometheus  address> [--upload-s3 <S3  destination>] [--region <AWS  region>] [--filename <Destination  file  name>]
```
  

Options

  
```
--query: (Required) The Prometheus query to fetch data.

--time-range: (Required) The time range for the query (e.g., "29d" for 29 days).

--address: (Required) The address of the Prometheus server.

--upload-s3: (Optional) The S3 destination in the format "s3://bucket-name/path/to/folder/". If provided, the CSV file will be uploaded to this S3 location.

--region: (Optional) The AWS region for S3. Required if --upload-s3 is specified.

--filename: (Optional) The name of the destination CSV file. If not provided, a default file name will be used.

--debug: (Optional) Enable debug output

```
All the following options can be used also via enviroment variables as you can see in the k8s/cronjob.tf example, the variables have the prefix `P2C_NAME` , here are some examples using environment variables:
  
```
P2C_QUERY='up' P2C_TIMERANGE='29d' P2C_ADDRESS='http://prometheus.ingress' P2C_UPLOAD_S3='s3://your-s3-buckets/metrics/' P2C_REGION='eu-west-1' ./p2c
```

## Examples

  

Export data from Prometheus to a local CSV file:


`./p2c --query "up" --time-range "7d" --address "http://prometheus.example.com:9090"`

  

Export data from Prometheus and upload it to an AWS S3 bucket:

`./p2c --query "up" --time-range "7d" --address "http://prometheus.example.com:9090" --upload-s3 "s3://my-s3-bucket/folder/" --region "us-west-1"`

  

Specify a custom destination file name:

`./p2c --query "up" --time-range "7d" --address "http://prometheus.example.com:9090" --filename "custom_data.csv"`

  

## Contributing

 
Contributions to this project are welcome! If you have any ideas, improvements, or bug fixes, feel free to create an issue or submit a pull request.

## License

  
This project is licensed under the MIT License - see the LICENSE file for details.
Feel free to customize the information as needed for your specific project.
