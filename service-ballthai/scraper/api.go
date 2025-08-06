package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// FetchAndParseAPI ดึงข้อมูลจาก URL ที่กำหนดและแปลงเป็นโครงสร้าง Go
// 'v' ควรเป็น pointer ไปยัง struct ที่ตรงกับโครงสร้าง JSON
func FetchAndParseAPI(url string, v interface{}) error {
	log.Printf("Fetching data from: %s", url)
	client := &http.Client{Timeout: 30 * time.Second} // ตั้งค่า Timeout
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("error fetching URL %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("API returned non-OK status %s from %s. Body: %s", resp.Status, url, string(bodyBytes))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body from %s: %w", url, err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("error unmarshalling JSON from %s: %w", url, err)
	}
	return nil
}

// DownloadImage ดาวน์โหลดรูปภาพจาก URL และบันทึกไปยังพาธที่กำหนด
// คืนค่าพาธที่บันทึกหรือ error
func DownloadImage(imageURL, saveDir string) (string, error) {
	if imageURL == "" {
		return "", fmt.Errorf("image URL cannot be empty")
	}

	// ดึงชื่อไฟล์จาก URL
	fileName := filepath.Base(imageURL)
	if fileName == "." || fileName == "/" { // จัดการกรณีที่ base อาจว่างเปล่าหรือเป็น root
		fileName = fmt.Sprintf("default_image_%d.png", time.Now().UnixNano()) // ชื่อไฟล์สำรอง
	}

	savePath := filepath.Join(saveDir, fileName)

	// ตรวจสอบว่าไดเรกทอรีมีอยู่หรือไม่ ถ้าไม่มีให้สร้าง
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		err = os.MkdirAll(saveDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", saveDir, err)
		}
	}

	// ตรวจสอบว่าไฟล์มีอยู่แล้วหรือไม่
	if _, err := os.Stat(savePath); err == nil {
		// log.Printf("Image already exists: %s", savePath)
		return savePath, nil // ถ้ามีอยู่แล้ว ไม่ต้องดาวน์โหลดซ้ำ
	}

	log.Printf("Downloading image from: %s to %s", imageURL, savePath)
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image from %s: %w", imageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image, status code: %d from %s", resp.StatusCode, imageURL)
	}

	out, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", savePath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write image to file %s: %w", savePath, err)
	}

	log.Printf("Image downloaded successfully: %s", savePath)
	return savePath, nil
}
