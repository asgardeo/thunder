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

package notification

import dbmodel "github.com/asgardeo/thunder/internal/system/database/model"

var (
	// queryCreateNotificationSender is the query to create a new notification sender.
	queryCreateNotificationSender = dbmodel.DBQuery{
		ID: "NMQ-SM-01",
		Query: "INSERT INTO NOTIFICATION_SENDER (NAME, SENDER_ID, DESCRIPTION, TYPE, PROVIDER) " +
			"VALUES ($1, $2, $3, $4, $5)",
	}

	// queryCreateNotificationSenderProperty is the query to create a notification sender property.
	queryCreateNotificationSenderProperty = dbmodel.DBQuery{
		ID: "NMQ-SM-02",
		Query: "INSERT INTO NOTIFICATION_SENDER_PROPERTY " +
			"(SENDER_ID, PROPERTY_NAME, PROPERTY_VALUE, IS_SECRET, IS_ENCRYPTED) " +
			"VALUES ($1, $2, $3, $4, $5)",
	}

	// queryGetNotificationSenderByID is the query to get a notification sender by its ID.
	queryGetNotificationSenderByID = dbmodel.DBQuery{
		ID:    "NMQ-SM-03",
		Query: "SELECT SENDER_ID, NAME, DESCRIPTION, TYPE, PROVIDER FROM NOTIFICATION_SENDER WHERE SENDER_ID = $1",
	}

	// queryGetNotificationSenderProperties is the query to get properties of a notification sender.
	queryGetNotificationSenderProperties = dbmodel.DBQuery{
		ID: "NMQ-SM-04",
		Query: "SELECT PROPERTY_NAME, PROPERTY_VALUE, IS_SECRET, IS_ENCRYPTED " +
			"FROM NOTIFICATION_SENDER_PROPERTY WHERE SENDER_ID = $1",
	}

	// queryGetAllNotificationSenders is the query to get all notification senders.
	queryGetAllNotificationSenders = dbmodel.DBQuery{
		ID:    "NMQ-SM-05",
		Query: "SELECT SENDER_ID, NAME, DESCRIPTION, TYPE, PROVIDER FROM NOTIFICATION_SENDER",
	}

	// queryUpdateNotificationSender is the query to update a notification sender.
	queryUpdateNotificationSender = dbmodel.DBQuery{
		ID: "NMQ-SM-06",
		Query: "UPDATE NOTIFICATION_SENDER SET NAME = $1, DESCRIPTION = $2, PROVIDER = $3, " +
			"UPDATED_AT = datetime('now') WHERE SENDER_ID = $4 AND TYPE = $5",
	}

	// queryDeleteNotificationSenderProperties is the query to delete all properties of a notification sender.
	queryDeleteNotificationSenderProperties = dbmodel.DBQuery{
		ID:    "NMQ-SM-07",
		Query: "DELETE FROM NOTIFICATION_SENDER_PROPERTY WHERE SENDER_ID = $1",
	}

	// queryDeleteNotificationSender is the query to delete a notification sender
	queryDeleteNotificationSender = dbmodel.DBQuery{
		ID:    "NMQ-SM-08",
		Query: "DELETE FROM NOTIFICATION_SENDER WHERE SENDER_ID = $1",
	}

	// queryGetNotificationSenderByName is the query to get a notification sender by name
	queryGetNotificationSenderByName = dbmodel.DBQuery{
		ID:    "NMQ-SM-09",
		Query: "SELECT SENDER_ID, NAME, DESCRIPTION, TYPE, PROVIDER FROM NOTIFICATION_SENDER WHERE NAME = $1",
	}
)
