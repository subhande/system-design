# Airline Checkin System

## Schema

### Trips Table

id | name


### Users Table

id | name


### Seats Table

id | name | trip_id | user_id


### Experiments

- Sequential assignment of seats
- Parallel assignment of seats without lock
- Parallel assignment of seats with lock
- Parallel assignment of seats with skip lock


## Ressult

Time taken to assign seats sequentially:  64.775042ms
Total seats assigned:  120


 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 

==================================================
Time taken to assign seats without lock:  30.4935ms
Total seats assigned:  9


 *  *  *  *  *  *  *  *  *  *  x  x  x  x  x  x  x  x  x  x 
 x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x 
 x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x 
 x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x 
 x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x 
 x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x 

==================================================
Time taken to assign seats with lock:  54.8835ms
Total seats assigned:  120


 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 

==================================================
Time taken to assign seats with skip lock:  24.381041ms
Total seats assigned:  120


 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 

==================================================