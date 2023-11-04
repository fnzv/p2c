package main

import (
    "context"
    "encoding/csv"
    "flag"
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/prometheus/client_golang/api"
    "github.com/prometheus/client_golang/api/prometheus/v1"
    "github.com/prometheus/common/model"
)

var s3Path string

func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        return defaultValue
    }
    return value
}

func main() {
    queryEnv := os.Getenv("P2C_QUERY")
    timeRangeEnv := os.Getenv("P2C_TIMERANGE")
    addressEnv := os.Getenv("P2C_ADDRESS")
    uploadS3Env := os.Getenv("P2C_UPLOAD_S3")
    regionEnv := os.Getenv("P2C_REGION")
    filenameEnv := os.Getenv("P2C_FILENAME")
    debugEnv := getEnv("P2C_DEBUG", "false")

    queryPtr := flag.String("query", queryEnv, "Prometheus query (can be specified multiple times, separated by spaces)")
    timeRangePtr := flag.String("time-range", timeRangeEnv, "Time range (e.g., 29d)")
    addressPtr := flag.String("address", addressEnv, "Prometheus address")
    uploadS3Ptr := flag.String("upload-s3", uploadS3Env, "S3 destination (e.g., s3://my-s3-bucket/folder/)")
    regionPtr := flag.String("region", regionEnv, "AWS region")
    filenamePtr := flag.String("filename", filenameEnv, "Destination file name")
    debugPtr := flag.String("debug", debugEnv, "Print debug information")

    flag.Parse()

    if *queryPtr == "" || *timeRangePtr == "" || *addressPtr == "" {
        fmt.Println("Usage: go run exporter.go --query <Prometheus query> --time-range <time range> --address <Prometheus address> [--upload-s3 <S3 destination>] [--region <AWS region>] [--filename <Destination file name>] [--debug]")
        os.Exit(1)
    }

    // Parse multiple queries from the command line
    queries := strings.Split(*queryPtr, " ")

    timeRangeStr := *timeRangePtr
    address := *addressPtr
    uploadS3 := *uploadS3Ptr
    region := *regionPtr
    filename := *filenamePtr
    debug := *debugPtr

    if debug == "true" {
        // Print debug information except AWS keys
        fmt.Printf("Debug information:\n")
        fmt.Printf("Address %s\n", address)
        fmt.Printf("Time Range %s\n", timeRangeStr)
        fmt.Printf("Upload S3 %s\n", uploadS3)
        fmt.Printf("Region %s\n", region)
        fmt.Printf("Filename %s\n", filename)
        for i, query := range queries {
            fmt.Printf("Query %d: %s\n", i+1, query)
        }
    }

    client, err := api.NewClient(api.Config{
        Address: address,
    })
    if err != nil {
        fmt.Printf("Error creating client: %v\n", err)
        os.Exit(1)
    }

    v1api := v1.NewAPI(client)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()


    // Parse the time range argument using regular expressions
    var start time.Time
    var end = time.Now()
    re := regexp.MustCompile(`^(\d+)([dDwWsSmM])$`)
    match := re.FindStringSubmatch(timeRangeStr)
    if match == nil || len(match) != 3 {
        fmt.Printf("Invalid time range format: %s\n", timeRangeStr)
        os.Exit(1)
    }

    value, unit := match[1], match[2]
    duration, err := parseTimeDuration(value, unit)
    if err != nil {
        fmt.Printf("Error parsing time range: %v\n", err)
        os.Exit(1)
    }

    start = end.Add(-duration)

    r := v1.Range{
        Start: start,
        End:   end,
        Step:  10 * time.Minute, // Adjust the step as needed
    }

    for i, query := range queries {
        result, warnings, err := v1api.QueryRange(ctx, query, r, v1.WithTimeout(5*time.Second))
        if err != nil {
            fmt.Printf("Error querying Prometheus (Query %d): %v\n", i+1, err)
            os.Exit(1)
        }
        if len(warnings) > 0 {
            fmt.Printf("Warnings (Query %d): %v\n", i+1, warnings)
        }

        // Prepare a CSV file for writing
        outputFile, err := os.Create(fmt.Sprintf("%s_%d.csv", filename, i+1))
        if err != nil {
            fmt.Printf("Error creating CSV file (Query %d): %v\n", i+1, err)
            os.Exit(1)
        }
        defer outputFile.Close()

        csvWriter := csv.NewWriter(outputFile)
        defer csvWriter.Flush()

        // Write the CSV header
        header := []string{"Value", "Timestamp"}
        if err := csvWriter.Write(header); err != nil {
            fmt.Printf("Error writing CSV header (Query %d): %v\n", i+1, err)
            os.Exit(1)
        }

        // Parse the result and format timestamps
        matrix, ok := result.(model.Matrix)
        if !ok {
            fmt.Printf("Error parsing the result into a Matrix (Query %d)\n", i+1)
            os.Exit(1)
        }

for _, sample := range matrix {
    for _, point := range sample.Values {
        timestamp := point.Timestamp.Time().Format("02/01/06 15:04:05")
        row := []string{fmt.Sprintf("%.3f", point.Value), timestamp} // Update formatting to %.3f
        if err := csvWriter.Write(row); err != nil {
            fmt.Printf("Error writing CSV row (Query %d): %v\n", i+1, err)
            os.Exit(1)
        }
    }
}

    }

    if uploadS3 != "" {
        for i, query := range queries {
            uploadToS3(uploadS3, fmt.Sprintf("%s_%d.csv", filename, i+1), region, queries[i])
            fmt.Println(query)
        }
    } else {
        for i := 1; i <= len(queries); i++ {
            fmt.Printf("File saved locally as %s_%d.csv\n", filename, i)
        }
    }
}

func parseTimeDuration(value, unit string) (time.Duration, error) {
    switch unit {
    case "s", "S":
        return time.Second * time.Duration(parseInt(value)), nil
    case "m", "M":
        return time.Minute * time.Duration(parseInt(value)), nil
    case "h", "H":
        return time.Hour * time.Duration(parseInt(value)), nil
    case "d", "D":
        return time.Hour * 24 * time.Duration(parseInt(value)), nil
    case "w", "W":
        return time.Hour * 24 * 7 * time.Duration(parseInt(value)), nil
    default:
        return 0, fmt.Errorf("invalid time unit: %s", unit)
    }
}

func parseInt(s string) int {
    val, _ := strconv.Atoi(s)
    return val
}




func uploadToS3(destination, filePath, region, queryName string) {
    currentDateTime := time.Now().Format("02-01-2006-15-04")
    destination = strings.Trim(destination, "s3://")
    s3URL := strings.TrimPrefix(destination, "s3://")
    parts := strings.SplitN(s3URL, "/", 2)

    if len(parts) != 2 {
        fmt.Println("Invalid S3 URL")
        return
    }

    bucketName := parts[0]
    s3Path := parts[1]

    cleanedQueryName := removeSpecialChars(queryName)
    objectKey := fmt.Sprintf("%s/%s-%s.csv", s3Path, currentDateTime, cleanedQueryName)

    sess, err := session.NewSessionWithOptions(session.Options{
        Config: aws.Config{
            Region: aws.String(region),
        },
        SharedConfigState: session.SharedConfigEnable,
    })
    if err != nil {
        fmt.Println("Error creating AWS session:", err)
        return
    }

    svc := s3.New(sess)

    file, err := os.Open(filePath)
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    fmt.Println("Upload to", objectKey)
    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(objectKey),
        Body:   file,
    })
    if err != nil {
        fmt.Println("Error uploading file to S3:", err)
        return
    }

    file, err = os.Open(filePath)
    if err != nil {
        fmt.Println("Error opening file for 'latest.csv' upload:", err)
        return
    }
    defer file.Close()

    objectKey = fmt.Sprintf("%s/latest-%s.csv", s3Path, cleanedQueryName)
    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(objectKey),
        Body:   file,
    })
    if err != nil {
        fmt.Println("Error uploading 'latest.csv' file to S3:", err)
        return
    }

    // Generate a pre-signed URL for the uploaded file
    req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(objectKey),
    })
    expiresIn := time.Hour * 24 * 7 // 7 days
    url, err := req.Presign(expiresIn)
    if err != nil {
        fmt.Println("Error generating pre-signed URL:", err)
        return
    }

    fmt.Println("Pre-signed URL:", url)
}

func removeSpecialChars(query string) string {
    // Remove special characters and spaces using regular expressions
    reg := regexp.MustCompile("[^a-zA-Z0-9]+")
    cleaned := reg.ReplaceAllString(query, "_")
    return cleaned
}
