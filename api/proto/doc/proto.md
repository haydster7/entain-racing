# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [racing/racing.proto](#racing_racing-proto)
    - [GetRaceRequest](#racing-GetRaceRequest)
    - [GetRaceResponse](#racing-GetRaceResponse)
    - [ListRacesRequest](#racing-ListRacesRequest)
    - [ListRacesRequestFilter](#racing-ListRacesRequestFilter)
    - [ListRacesResponse](#racing-ListRacesResponse)
    - [Race](#racing-Race)
  
    - [Racing](#racing-Racing)
  
- [sports/sports.proto](#sports_sports-proto)
    - [Event](#sports-Event)
    - [GetEventRequest](#sports-GetEventRequest)
    - [GetEventResponse](#sports-GetEventResponse)
    - [ListEventsRequest](#sports-ListEventsRequest)
    - [ListEventsRequestFilter](#sports-ListEventsRequestFilter)
    - [ListEventsResponse](#sports-ListEventsResponse)
  
    - [Sports](#sports-Sports)
  
- [Scalar Value Types](#scalar-value-types)



<a name="racing_racing-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## racing/racing.proto



<a name="racing-GetRaceRequest"></a>

### GetRaceRequest
Request to GetRace call


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  |  |






<a name="racing-GetRaceResponse"></a>

### GetRaceResponse
Response to GetRace call.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| race | [Race](#racing-Race) |  |  |






<a name="racing-ListRacesRequest"></a>

### ListRacesRequest
Request to ListRaces call


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filter | [ListRacesRequestFilter](#racing-ListRacesRequestFilter) |  |  |
| order_by | [string](#string) |  |  |






<a name="racing-ListRacesRequestFilter"></a>

### ListRacesRequestFilter
Filter for listing races.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| meeting_ids | [int64](#int64) | repeated |  |
| visible | [bool](#bool) | optional |  |






<a name="racing-ListRacesResponse"></a>

### ListRacesResponse
Response to ListRaces call.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| races | [Race](#racing-Race) | repeated |  |






<a name="racing-Race"></a>

### Race
A race resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  | ID represents a unique identifier for the race. |
| meeting_id | [int64](#int64) |  | MeetingID represents a unique identifier for the races meeting. |
| name | [string](#string) |  | Name is the official name given to the race. |
| number | [int64](#int64) |  | Number represents the number of the race. |
| visible | [bool](#bool) |  | Visible represents whether or not the race is visible. |
| advertised_start_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | AdvertisedStartTime is the time the race is advertised to run. |
| status | [string](#string) |  | status determines if a race is open or closed based on advertised_start_time |





 

 

 


<a name="racing-Racing"></a>

### Racing


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListRaces | [ListRacesRequest](#racing-ListRacesRequest) | [ListRacesResponse](#racing-ListRacesResponse) | ListRaces returns a list of all races. |
| GetRace | [GetRaceRequest](#racing-GetRaceRequest) | [GetRaceResponse](#racing-GetRaceResponse) | GetRace returns a single race matching the requested id |

 



<a name="sports_sports-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## sports/sports.proto



<a name="sports-Event"></a>

### Event
A sport event resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  | ID represents a unique identifier for the event. |
| home_team | [string](#string) |  | Name of the team playing on home grounds. |
| away_team | [string](#string) |  | Name of the visitng team. |
| sport | [string](#string) |  | Name of the sport being played. |
| location | [string](#string) |  | City where the event is being held. |
| capacity | [int64](#int64) |  | Capacity of venue the event is being held. |
| advertised_start_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | AdvertisedStartTime is the time the event is advertised to run. |
| expected_end_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | ExpectedEndTime is the time the event is expected to end with no overtime. |
| status | [string](#string) |  | Status determines if a sport is open, in progress, or closed based on start and end times |






<a name="sports-GetEventRequest"></a>

### GetEventRequest
Request to GetEvent call


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  |  |






<a name="sports-GetEventResponse"></a>

### GetEventResponse
Response to GetSport call.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#sports-Event) |  |  |






<a name="sports-ListEventsRequest"></a>

### ListEventsRequest
Request to ListEvents call


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filter | [ListEventsRequestFilter](#sports-ListEventsRequestFilter) |  |  |
| order_by | [string](#string) |  |  |






<a name="sports-ListEventsRequestFilter"></a>

### ListEventsRequestFilter
Filter for listing sport events


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sport | [string](#string) |  |  |
| team | [string](#string) |  |  |
| status | [string](#string) |  |  |






<a name="sports-ListEventsResponse"></a>

### ListEventsResponse
Response to ListEvents call.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| events | [Event](#sports-Event) | repeated |  |





 

 

 


<a name="sports-Sports"></a>

### Sports


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListEvents | [ListEventsRequest](#sports-ListEventsRequest) | [ListEventsResponse](#sports-ListEventsResponse) | ListSports will return a collection of sport events |
| GetEvent | [GetEventRequest](#sports-GetEventRequest) | [GetEventResponse](#sports-GetEventResponse) | GetSport will return a single sport event matching the requested id |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

