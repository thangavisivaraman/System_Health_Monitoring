package system;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Statement;
import java.util.ArrayList;
import java.util.List;

import org.json.JSONArray;
import org.json.JSONObject;

import com.google.gson.Gson;
import com.google.gson.JsonObject;

public class DBFunction {

	public static void insertData(String key, JSONObject systemInfo) {
		Connection connection = DBConnection.getInstance().getConnection();

		try {
			// Step 1: Check if the table exists
			if (!tableExists(connection, key)) {
				// Step 2: Create the table if it does not exist
				// createTable(connection, key);
				createTable(connection, key);
			}

			// Step 3: Insert the JSON data into the table
			insertJsonData(connection, key, systemInfo);

		} catch (SQLException e) {
			e.printStackTrace();
		}

	}

	private static void createTable(Connection connection, String tableName) throws SQLException {
		// Create a table with id and context (to store JSON data)
		String createTableSQL = String.format("CREATE TABLE `%s` (" + "`id` INT NOT NULL AUTO_INCREMENT,"
				+ "`context` TEXT NOT NULL," + "PRIMARY KEY (`id`)" + ")", tableName);
		try (Statement stmt = connection.createStatement()) {
			stmt.execute(createTableSQL);
		}
	}

	private static boolean tableExists(Connection connection, String tableName) throws SQLException {
		// Check if the table exists
		String query = "SHOW TABLES LIKE ?";
		try (PreparedStatement stmt = connection.prepareStatement(query)) {
			stmt.setString(1, tableName);
			ResultSet rs = stmt.executeQuery();
			return rs.next(); // Returns true if the table exists
		}
	}



	private static void insertJsonData(Connection connection, String tableName, JSONObject systemInfo)
			throws SQLException {
		// Insert the JSON data into the context column
		String insertSQL = String.format("INSERT INTO `%s` (context) VALUES (?)", tableName);

		try (PreparedStatement pstmt = connection.prepareStatement(insertSQL)) {
			// Convert JSONObject to string and store it in the context column
			pstmt.setString(1, systemInfo.toString());
			pstmt.executeUpdate();
		}
	}

	public static List<JsonObject> getDataWithLimitAndOffset(String tableName, int limit, int offset) {
		
		Gson gson = new Gson(); // Initialize Gson instance
		List<JsonObject> dataList = new ArrayList<>(); // List to store JsonObjects
		
		Connection connection = DBConnection.getInstance().getConnection();

		try {
		
			// Prepare SQL query with limit and offset
			 String query = "SELECT * FROM " + tableName + " ORDER BY id DESC LIMIT ? OFFSET ?";
			PreparedStatement pstmt = connection.prepareStatement(query);
			pstmt.setInt(1, limit);
			pstmt.setInt(2, offset);
			ResultSet rs = pstmt.executeQuery();

			// Process the result set and build JSON response
			while (rs.next()) {
			
				String context = rs.getString("context");

				// Convert the context string to a JsonObject using Gson
				JsonObject jsonObject = gson.fromJson(context, JsonObject.class);

				// Add the JsonObject to the list
				dataList.add(jsonObject);
			}

			// Close the result set and statement
			rs.close();
			pstmt.close();

		} catch (SQLException e) {
			e.printStackTrace(); // Log SQL exceptions
		}

		System.out.println(dataList);

		// Return the results list
		return dataList;
	}

//	public static void main(String[] args) {
//		DBFunction.getDataWithLimitAndOffset("Thulasi_info", 20, 0);
//	}
		public static  boolean authenticateUser(String username, String password) throws Exception {
        // Connect to database and check if the user exists
		
		
		
        try  
        {
        	Connection connection = DBConnection.getInstance().getConnection();
            String query = "SELECT username FROM users WHERE username = ? AND password = ?";
            PreparedStatement stmt = connection.prepareStatement(query);
            stmt.setString(1, username);
            stmt.setString(2, password); // Do NOT store plain text passwords in production!
            ResultSet rs = stmt.executeQuery();

            return rs.next(); // Return true if a match is found
        }catch(Exception e) {
        	System.out.println(e);
        }
		return false;
        
    }

}
