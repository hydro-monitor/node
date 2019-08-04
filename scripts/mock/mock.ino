void setup() {
    // Turn the Serial Protocol ON
    Serial.begin(9600);
}

void loop() {
    // check if data has been sent from the computer
    if (Serial.available()) {
        byte byteRead;
        byteRead = Serial.read();
        // Send water level
        Serial.println("432");
    }
}
