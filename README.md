#  CSV2API

####  CSV2API is a web service that converts CSV files to RESTful APIs
**Features**
 - Accept multiple user uploaded CSV files
 - Concurrently processes files
 - Interact with CSV data through RESTful API
 - Authentication system with Registration/Login, Session Token, and API key
 - Data stored in PostgreSQL database and user session information stored in Redis cache
 



|          Function          | Method |                              Route                              |      Auth     |
|:--------------------------:|:------:|:---------------------------------------------------------------:|:-------------:|
|          Register          |  POST  | /register                                                       |       No      |
|            Login           |  POST  | /login                                                          |       No      |
|        Upload Files        |  POST  | /upload                                                         | Session Token |
|      Get All Documents     |   GET  | /{username}/documents                                           |    API Key    |
|     Get Single Document    |   GET  | /{username}/documents/{id}                                      |    API Key    |
|      Delete Document     | DELETE | /{username}/documents/{id}                                      |    API Key    |
|  Get All Rows In Document  |   GET  | /{username}/documents/{docID}/rows                              |    API Key    |
| Create Row In Document |  POST  | /{username}/documents/{docID}/rows                              |    API Key    |
|   Get Row In Document  |   GET  | /{username}/documents/{docID}/rows/{rowID}                      |    API Key    |
|  Update Row In Document  |   PUT  | /{username}/documents/{docID}/rows/{rowID}                      |    API Key    |
|  Delete Row In Document  | DELETE | /{username}/documents/{docID}/rows/{rowID}                      |    API Key    |
|   Get Rows By Parameters   |   GET  | /{username}/documents/{docID}/rows?column={columns}&data={data} |    API Key    |
