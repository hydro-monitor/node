void setup() {
    // Turn the Serial Protocol ON
    Serial.begin(9600);
}

const char* measurements[] = {"-42", "2", "2", "44", "55", "1", "1", "-5", "-45"};
int lenMeasurements = 9;
int i = 0;

void loop() {
    // check if data has been sent from the computer
    if (Serial.available()) {
        byte byteRead;
        byteRead = Serial.read();
        // Send water level
        Serial.println(measurements[i % lenMeasurements]);
        i++;
    }
}
