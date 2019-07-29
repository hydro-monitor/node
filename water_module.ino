//Biblioteca para tarjeta SD
#include <SD.h>

//Biblioteca para sensor ultrasonico HC-SR04
#include <NewPing.h>

//Biblioteca para timer
#include <Wire.h>
#include "RTClib.h"

const int chipSelect = 10; //Chip select para la SD

#define SENSOR_DISTANCE 600  // Altura a la cual se encuentra el sensor
#define INTERVALO_MEDICION 2 //Intervalo de tiempo (en segundos) en el que se toma una medicion

//Pines asignados para el ultrasonido
#define TRIGGER_PIN  9       // Disparo que habilita el sensado     
#define ECHO_PIN     8       // Señal de salida del sensor que contiene la informacion sobre la distancia detectada
#define MAX_DISTANCE 400     // Máxima distancia que puede medir el sensor
#define ITERACIONES 5        // Cantidad de iteraciones para el promediado de datos

unsigned long distancia;
unsigned long tiempo;

NewPing sonar(TRIGGER_PIN, ECHO_PIN, MAX_DISTANCE);

RTC_DS1307 rtc;

void setup()
{
 
  //Inicializa la placa del timer. Ya esta puesta en hora
  Wire.begin();
  rtc.begin();
  
  //Inicializa la tarjeta SD
  pinMode(10, OUTPUT); //El pin de chip select por omision tiene que estar como salida aunque no sea el que se use
  
  //Veo si la tarjeta esta presente y puede ser inicializada:
  if (!SD.begin(chipSelect)) {
    //si no está, no hace nada mas
    return;
  }

  //La siguiente linea pone en hora el RTC al día y la fecha en la que se compila este codigo
  //rtc.adjust(DateTime(F(__DATE__), F(__TIME__)));
   
  //Veo si el RTC esta corriendo 
  if (! rtc.isrunning()) {
    return;
  }
  
   
}

void loop()
{
  //Cadena para guardar la medición
  String stringNivel = ""; 
  
  //Cadenas para armar la fecha y hora
  String stringFecha = ""; 
  String stringAnio = "";
  String stringMes = ""; 
  String stringDia = "";
  String stringHora = "";
  String stringMinuto = ""; 

  //Cadena "fecha, medicion". Es la que se guardara en la SD
  String stringDatos = "";
 
  //Variables para la medicion
  int tiempo;
  int distancia; 
  int nivel;
  
  int retardo; //Para calcular los milisegundos para pasarle a la funcion delay
  
  // Se realiza un promedio de los valores registrados, descartando los valores fuera de rango
  tiempo = sonar.ping_median(ITERACIONES);
  // Se convierte el valor de tiempo a cm
  distancia = sonar.convert_cm(tiempo);
 
  // Calculo del nivel detectado
  nivel = SENSOR_DISTANCE - distancia;

  //Guarda la medicion en la cadena
  stringNivel=String(nivel);
  
  //Levanta la fecha
  DateTime now = rtc.now();
  //Arma las cadenenas para la fecha
  stringAnio=String(now.year(), DEC);
  stringMes=String(now.month(), DEC);
  stringDia=String(now.day(), DEC);
  stringHora=String(now.hour(), DEC);
  stringMinuto=String(now.minute(), DEC);
  stringFecha= stringAnio + "-" + stringMes + "-" + stringDia + "-" + stringHora + "-" + stringMinuto;

  //Arma la cadena a guardar
  stringDatos=stringFecha + "," + stringNivel;
  
  //Abre el archivo. Se puede abrir un solo archivo por vez
  File dataFile = SD.open("datalog.txt", FILE_WRITE);

  //Si el archivo esta disponible, escribe en el
  if (dataFile) {
    dataFile.println(stringDatos);
    dataFile.close();
  }  
    
    retardo=INTERVALO_MEDICION*1000;
    delay(retardo);
}

