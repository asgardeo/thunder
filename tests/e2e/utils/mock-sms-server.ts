/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import express, { Express, Request, Response } from "express";
import { Server } from "http";

/**
 * Represents a captured SMS message with extracted OTP
 */
export interface SMSMessage {
  message: string;
  otp: string;
  timestamp: Date;
}

/**
 * Mock SMS Server for E2E Testing
 *
 * This server acts as a fake SMS provider that captures messages sent by Thunder
 * during authentication flows, automatically extracts OTP codes, and provides
 * endpoints for tests to retrieve the captured messages.
 *
 * Features:
 * - POST /send-sms - Endpoint for Thunder to send SMS messages
 * - GET /messages - Retrieve all captured messages
 * - GET /messages/last - Get the most recent message
 * - POST /clear - Clear all captured messages
 * - Automatic OTP extraction from message body
 *
 * @example
 * ```typescript
 * const mockServer = new MockSMSServer(8098);
 * await mockServer.start();
 *
 * // Later in test...
 * const lastMessage = mockServer.getLastMessage();
 * const otp = lastMessage?.otp;
 *
 * await mockServer.stop();
 * ```
 */
export class MockSMSServer {
  private app: Express;
  private server: Server | null = null;
  private messages: SMSMessage[] = [];
  private port: number;

  constructor(port: number = 8098) {
    this.port = port;
    this.app = express();
    this.setupMiddleware();
    this.setupRoutes();
  }

  /**
   * Configure Express middleware
   */
  private setupMiddleware(): void {
    // Accept all content types as plain text - SMS messages are always text
    this.app.use(express.text({ type: "*/*" }));
  }

  /**
   * Setup HTTP routes
   */
  private setupRoutes(): void {
    // Main endpoint that Thunder calls to send SMS
    this.app.post("/send-sms", this.handleSendSMS.bind(this));

    // Endpoint for tests to retrieve all messages
    this.app.get("/messages", this.handleGetMessages.bind(this));

    // Endpoint for tests to get last message
    this.app.get("/messages/last", this.handleGetLastMessage.bind(this));

    // Endpoint to clear all messages
    this.app.post("/clear", this.handleClearMessages.bind(this));

    // Health check
    this.app.get("/health", (_req: Request, res: Response) => {
      res.json({ status: "ok", messagesCount: this.messages.length });
    });
  }

  /**
   * Handle incoming SMS from Thunder
   */
  private handleSendSMS(req: Request, res: Response): void {
    try {
      // SMS message is always plain text
      const messageBody = typeof req.body === "string" ? req.body : String(req.body || "");

      // Extract OTP from message
      const otp = this.extractOTP(messageBody);

      // Store message
      const smsMessage: SMSMessage = {
        message: messageBody,
        otp: otp,
        timestamp: new Date(),
      };

      this.messages.push(smsMessage);

      console.log(
        `[Mock SMS Server] Message received: "${messageBody.substring(0, 50)}${messageBody.length > 50 ? "..." : ""}" | OTP: ${otp || "none"}`
      );

      // Return success response
      res.status(200).json({
        success: true,
        messageId: `mock-msg-${this.messages.length}`,
        timestamp: smsMessage.timestamp.toISOString(),
      });
    } catch (error) {
      console.error("[Mock SMS Server] Error handling SMS:", error);
      res.status(500).json({
        success: false,
        error: "Failed to process SMS message",
      });
    }
  }

  /**
   * Handle request to retrieve all messages
   */
  private handleGetMessages(_req: Request, res: Response): void {
    res.json({
      count: this.messages.length,
      messages: this.messages,
    });
  }

  /**
   * Handle request to get last message
   */
  private handleGetLastMessage(_req: Request, res: Response): void {
    const lastMessage = this.messages.length > 0 ? this.messages[this.messages.length - 1] : null;
    res.json(lastMessage);
  }

  /**
   * Handle request to clear all messages
   */
  private handleClearMessages(_req: Request, res: Response): void {
    const clearedCount = this.messages.length;
    this.messages = [];
    console.log(`[Mock SMS Server] Cleared ${clearedCount} message(s)`);
    res.json({
      status: "cleared",
      count: clearedCount,
    });
  }

  /**
   * Extract OTP from SMS message body
   *
   * Handles patterns like: "Your verification code is: 841317. This code will..."
   * Searches for numeric sequences between 4-8 digits and returns the most
   * likely OTP code, prioritizing 6-digit codes (most common for SMS OTP).
   *
   * @param message - The SMS message body
   * @returns Extracted OTP code or empty string if none found
   */
  private extractOTP(message: string): string {
    if (!message) return "";

    // First, try to extract OTP from common patterns
    // Pattern: "Your verification code is: 123456" or "code is: 123456" or "code: 123456"
    const patternMatch = message.match(/(?:verification code|code)\s*(?:is\s*)?:\s*(\d{4,8})/i);
    if (patternMatch && patternMatch[1]) {
      return patternMatch[1];
    }

    // Fallback: Find all numeric sequences of 4-8 digits
    const matches = message.match(/\b\d{4,8}\b/g);

    if (!matches || matches.length === 0) {
      return "";
    }

    // Score each potential OTP
    const scored = matches.map(match => ({
      value: match,
      score: this.calculateOTPScore(match),
    }));

    // Sort by score (highest first)
    scored.sort((a, b) => b.score - a.score);

    return scored[0].value;
  }

  /**
   * Calculate score for potential OTP sequence
   *
   * Prioritizes:
   * 1. 6-digit codes (most common) - score 100
   * 2. 4-digit codes - score 80
   * 3. 5-digit codes - score 70
   * 4. 8-digit codes - score 60
   * 5. 7-digit codes - score 50
   *
   * @param sequence - Numeric sequence to score
   * @returns Score value
   */
  private calculateOTPScore(sequence: string): number {
    const length = sequence.length;

    switch (length) {
      case 6:
        return 100; // Most common OTP length
      case 4:
        return 80;
      case 5:
        return 70;
      case 8:
        return 60;
      case 7:
        return 50;
      default:
        return 0;
    }
  }

  /**
   * Start the mock SMS server
   *
   * @returns Promise that resolves when server is listening
   */
  async start(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.server = this.app.listen(this.port, () => {
          console.log(`[Mock SMS Server] Started on http://localhost:${this.port}`);
          console.log(`[Mock SMS Server] SMS endpoint: http://localhost:${this.port}/send-sms`);
          resolve();
        });

        this.server.on("error", (error: Error) => {
          console.error(`[Mock SMS Server] Failed to start:`, error);
          reject(error);
        });
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Stop the mock SMS server
   *
   * @returns Promise that resolves when server is closed
   */
  async stop(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.server) {
        resolve();
        return;
      }

      this.server.close(err => {
        if (err) {
          console.error("[Mock SMS Server] Error stopping server:", err);
          reject(err);
        } else {
          console.log("[Mock SMS Server] Stopped");
          this.server = null;
          resolve();
        }
      });
    });
  }

  /**
   * Get the SMS sending endpoint URL
   *
   * @returns Full URL to the /send-sms endpoint
   */
  getSendSMSURL(): string {
    return `http://localhost:${this.port}/send-sms`;
  }

  /**
   * Get base URL of the mock server
   *
   * @returns Base URL
   */
  getURL(): string {
    return `http://localhost:${this.port}`;
  }

  /**
   * Get the last received message
   *
   * @returns Last SMS message or null if no messages
   */
  getLastMessage(): SMSMessage | null {
    return this.messages.length > 0 ? this.messages[this.messages.length - 1] : null;
  }

  /**
   * Get all received messages
   *
   * @returns Array of all SMS messages
   */
  getAllMessages(): SMSMessage[] {
    return [...this.messages];
  }

  /**
   * Clear all stored messages
   */
  clearMessages(): void {
    this.messages = [];
  }

  /**
   * Get count of received messages
   *
   * @returns Number of messages
   */
  getMessageCount(): number {
    return this.messages.length;
  }

  /**
   * Check if server is running
   *
   * @returns true if server is listening
   */
  isRunning(): boolean {
    return this.server !== null && this.server.listening;
  }
}
