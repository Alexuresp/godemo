package com.example.springapp;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class HealthController {

  @Value("${APP_NAME:spring-app}")
  private String appName;

  @Value("${GREETING:Hello}")
  private String greeting;

  @GetMapping("/actuator/health")
  public ResponseEntity<String> health() {
    return ResponseEntity.ok("OK");
  }

  @GetMapping("/")
  public ResponseEntity<String> index() {
    return ResponseEntity.ok(greeting + " from " + appName);
  }
}
