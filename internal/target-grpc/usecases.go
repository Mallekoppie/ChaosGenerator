package targetgrpc

import (
	"fmt"
	"strings"
	"time"

	"mallekoppie/ChaosGenerator/internal/contracts"
)

// GenerateLargePayload generates a byte array of specified size
func GenerateLargePayload(sizeBytes int) []byte {
	data := make([]byte, sizeBytes)
	// Fill with some pattern for realistic data
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// GenerateBulkUpdateOperations generates bulk update operations
func GenerateBulkUpdateOperations(count int) []*contracts.UpdateOperation {
	operations := make([]*contracts.UpdateOperation, count)
	for i := 0; i < count; i++ {
		operations[i] = &contracts.UpdateOperation{
			Id: fmt.Sprintf("user-%d", i),
			Fields: map[string]string{
				"lastLogin":         time.Now().Format(time.RFC3339),
				"preferences.theme": "dark",
			},
		}
	}
	return operations
}

// GenerateUserSummaries generates a list of user summaries
func GenerateUserSummaries(count int, startIndex int) []*contracts.UserSummary {
	users := make([]*contracts.UserSummary, count)
	for i := 0; i < count; i++ {
		idx := startIndex + i
		users[i] = &contracts.UserSummary{
			Id:        fmt.Sprintf("user-%d", idx),
			Username:  fmt.Sprintf("user%d", idx),
			Email:     fmt.Sprintf("user%d@example.com", idx),
			FirstName: "User",
			LastName:  fmt.Sprintf("%d", idx),
		}
	}
	return users
}

// GenerateReportData generates report data points
func GenerateReportData(days int) []*contracts.ReportDataPoint {
	data := make([]*contracts.ReportDataPoint, days)
	baseDate := time.Now().AddDate(0, 0, -days)

	for i := 0; i < days; i++ {
		date := baseDate.AddDate(0, 0, i)
		data[i] = &contracts.ReportDataPoint{
			Date:         date.Format("2006-01-02"),
			Category:     "retail",
			Transactions: int32(100 + i*10),
			Amount:       10000.00 + float64(i)*500.50,
			Details:      GenerateLargePayload(1024), // 1KB of details per day
		}
	}
	return data
}

// GenerateTransactionRecords generates transaction records for export
func GenerateTransactionRecords(count int) []*contracts.TransactionRecord {
	records := make([]*contracts.TransactionRecord, count)
	baseTime := time.Now().AddDate(0, 0, -30)

	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		records[i] = &contracts.TransactionRecord{
			TransactionId: fmt.Sprintf("txn-%d", i),
			Timestamp:     timestamp.Format(time.RFC3339),
			Amount:        99.99 + float64(i%100),
			Currency:      "USD",
			Merchant:      fmt.Sprintf("Store-%d", i%10),
			Category:      "retail",
			Status:        "completed",
			Metadata:      GenerateLargePayload(200), // 200 bytes metadata
		}
	}
	return records
}

// GenerateDataRecords generates sensor data records for processing
func GenerateDataRecords(count int) []*contracts.DataRecord {
	records := make([]*contracts.DataRecord, count)
	baseTime := time.Now()

	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Second)
		records[i] = &contracts.DataRecord{
			Timestamp: timestamp.Format(time.RFC3339Nano),
			SensorId:  fmt.Sprintf("sensor-%d", i%100),
			Values: &contracts.SensorValues{
				Temperature:    22.5 + float64(i%10),
				Humidity:       45.2 + float64(i%20),
				Pressure:       1013.25 + float64(i%5),
				WindSpeed:      12.3 + float64(i%15),
				WindDirection:  float64(i % 360),
				Precipitation:  0.0,
				SolarRadiation: 450.5 + float64(i%100),
				UvIndex:        3.2 + float64(i%8),
				AirQuality:     85.0 + float64(i%15),
				NoiseLevel:     45.6 + float64(i%30),
			},
			Metadata: &contracts.RecordMetadata{
				Location: &contracts.Location{
					Lat: 40.7128 + float64(i%10)/100,
					Lon: -74.0060 + float64(i%10)/100,
				},
				Quality:    "high",
				Calibrated: true,
			},
		}
	}
	return records
}

// CalculateStatistics calculates min/max/avg/stddev for sensor data
func CalculateStatistics(records []*contracts.DataRecord) *contracts.Statistics {
	if len(records) == 0 {
		return &contracts.Statistics{
			Min:     &contracts.SensorValues{},
			Max:     &contracts.SensorValues{},
			Average: &contracts.SensorValues{},
			Stddev:  &contracts.SensorValues{},
		}
	}

	// Initialize with first record's values
	min := &contracts.SensorValues{
		Temperature:    records[0].Values.Temperature,
		Humidity:       records[0].Values.Humidity,
		Pressure:       records[0].Values.Pressure,
		WindSpeed:      records[0].Values.WindSpeed,
		WindDirection:  records[0].Values.WindDirection,
		Precipitation:  records[0].Values.Precipitation,
		SolarRadiation: records[0].Values.SolarRadiation,
		UvIndex:        records[0].Values.UvIndex,
		AirQuality:     records[0].Values.AirQuality,
		NoiseLevel:     records[0].Values.NoiseLevel,
	}
	max := &contracts.SensorValues{
		Temperature:    records[0].Values.Temperature,
		Humidity:       records[0].Values.Humidity,
		Pressure:       records[0].Values.Pressure,
		WindSpeed:      records[0].Values.WindSpeed,
		WindDirection:  records[0].Values.WindDirection,
		Precipitation:  records[0].Values.Precipitation,
		SolarRadiation: records[0].Values.SolarRadiation,
		UvIndex:        records[0].Values.UvIndex,
		AirQuality:     records[0].Values.AirQuality,
		NoiseLevel:     records[0].Values.NoiseLevel,
	}
	sum := &contracts.SensorValues{
		Temperature:    records[0].Values.Temperature,
		Humidity:       records[0].Values.Humidity,
		Pressure:       records[0].Values.Pressure,
		WindSpeed:      records[0].Values.WindSpeed,
		WindDirection:  records[0].Values.WindDirection,
		Precipitation:  records[0].Values.Precipitation,
		SolarRadiation: records[0].Values.SolarRadiation,
		UvIndex:        records[0].Values.UvIndex,
		AirQuality:     records[0].Values.AirQuality,
		NoiseLevel:     records[0].Values.NoiseLevel,
	}

	// Calculate min, max, sum
	for i := 1; i < len(records); i++ {
		v := records[i].Values

		if v.Temperature < min.Temperature {
			min.Temperature = v.Temperature
		}
		if v.Temperature > max.Temperature {
			max.Temperature = v.Temperature
		}
		sum.Temperature += v.Temperature

		if v.Humidity < min.Humidity {
			min.Humidity = v.Humidity
		}
		if v.Humidity > max.Humidity {
			max.Humidity = v.Humidity
		}
		sum.Humidity += v.Humidity

		if v.Pressure < min.Pressure {
			min.Pressure = v.Pressure
		}
		if v.Pressure > max.Pressure {
			max.Pressure = v.Pressure
		}
		sum.Pressure += v.Pressure
	}

	// Calculate averages
	count := float64(len(records))
	avg := &contracts.SensorValues{
		Temperature: sum.Temperature / count,
		Humidity:    sum.Humidity / count,
		Pressure:    sum.Pressure / count,
		WindSpeed:   sum.WindSpeed / count,
	}

	// For simplicity, stddev is calculated as a percentage of average
	stddev := &contracts.SensorValues{
		Temperature: avg.Temperature * 0.1,
		Humidity:    avg.Humidity * 0.1,
		Pressure:    avg.Pressure * 0.05,
	}

	return &contracts.Statistics{
		Min:     min,
		Max:     max,
		Average: avg,
		Stddev:  stddev,
	}
}

// GenerateBulkUpdatePayload generates a bulk update payload (~100KB)
func GenerateBulkUpdatePayload() string {
	updates := make([]string, 100)
	for i := 0; i < 100; i++ {
		updates[i] = fmt.Sprintf(`{"id":"user-%d","fields":{"lastLogin":"2026-02-14T10:30:00Z","preferences.theme":"dark"}}`, i)
	}
	return fmt.Sprintf(`{"updates":[%s],"options":{"validateAll":true,"atomic":false}}`, strings.Join(updates, ","))
}
