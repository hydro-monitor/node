#include <NewPing.h>			// Biblioteca para sensor ultrasonico HC-SR04

#define SENSOR_DISTANCE 600		// Altura a la cual se encuentra el sensor, medida desde el fondo del río
#define MAX_DISTANCE 400		// Máxima distancia que puede medir el sensor expresada en cm
#define ITERACIONES 5			// Cantidad de iteraciones para el promediado de datos

// Pines asignados para el ultrasonido

#define TRIGGER_PIN_1  7		// Disparo que habilita el sensado para el sensor 1
#define ECHO_PIN_1     6		// Señal de salida del sensor que contiene la informacion sobre la distancia detectada para el sensor 1
#define TRIGGER_PIN_2  5		// Disparo que habilita el sensado para el sensor 2
#define ECHO_PIN_2     4		// Señal de salida del sensor que contiene la informacion sobre la distancia detectada para el sensor 2
#define TRIGGER_PIN_3  3		// Disparo que habilita el sensado para el sensor 3
#define ECHO_PIN_3     2		// Señal de salida del sensor que contiene la informacion sobre la distancia detectada para el sensor 3

// Inicialización de los sensores

NewPing sonar_1(TRIGGER_PIN_1, ECHO_PIN_1, MAX_DISTANCE);
NewPing sonar_2(TRIGGER_PIN_2, ECHO_PIN_2, MAX_DISTANCE);
NewPing sonar_3(TRIGGER_PIN_3, ECHO_PIN_3, MAX_DISTANCE);

void setup() {
	Serial.begin(9600);	// Inicialización del puerto serie a 9600 bps
}

void loop() {
	if(Serial.available()) {
		// Recibir signal de medicion
		byte byte_read;
        byte_read = Serial.read();

  		// Variables para la medicion
	  	int tiempo_sensor_1, tiempo_sensor_2, tiempo_sensor_3;
		int distancia_sensor_1, distancia_sensor_2, distancia_sensor_3; 
		int nivel;
	    
		// Se realiza un promedio de los valores registrados, descartando los valores fuera de rango
		tiempo_sensor_1 = sonar_1.ping_median(ITERACIONES);
		tiempo_sensor_2 = sonar_2.ping_median(ITERACIONES);
		tiempo_sensor_3 = sonar_3.ping_median(ITERACIONES);
		// Se convierte el valor de tiempo a cm
		distancia_sensor_1 = sonar_1.convert_cm(tiempo_sensor_1);
		distancia_sensor_2 = sonar_2.convert_cm(tiempo_sensor_2);
		distancia_sensor_3 = sonar_3.convert_cm(tiempo_sensor_3);
	 
		// Calculo del nivel detectado en base a una verificación de 2 de 3
		if (distancia_sensor_1 == distancia_sensor_2) {
			nivel = SENSOR_DISTANCE - distancia_sensor_1;
			Serial.println(nivel);
			return
		} else if (distancia_sensor_1 == distancia_sensor_3) {
			nivel = SENSOR_DISTANCE - distancia_sensor_1;
			Serial.println(nivel);
			return
		} else if (distancia_sensor_2 == distancia_sensor_3) {
			nivel = SENSOR_DISTANCE - distancia_sensor_2;
			Serial.println(nivel);
			return
		}

		// No hay coincidencia entre mediciones
		Serial.println("Inconcluso");
	}
}