package system;

import jakarta.servlet.ServletException;
import jakarta.servlet.annotation.WebServlet;
import jakarta.servlet.http.HttpServlet;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.PrintWriter;

import org.json.JSONObject;

/**
 * Servlet implementation class LoginServlet
 */
@WebServlet("/LoginServlet")
public class LoginServlet extends HttpServlet {
	private static final long serialVersionUID = 1L;
       
    public LoginServlet() {
        super();
        // TODO Auto-generated constructor stub
    }

	protected void doGet(HttpServletRequest request, HttpServletResponse response) throws ServletException, IOException {
		// TODO Auto-generated method stub
		response.getWriter().append("Served at: ").append(request.getContextPath());
	}

	@Override
    protected void doPost(HttpServletRequest request, HttpServletResponse response) throws ServletException, IOException {
        response.setContentType("application/json");
        PrintWriter out = response.getWriter();
        JSONObject jsonResponse = new JSONObject();

        try {
            // Parse JSON request body
            StringBuilder sb = new StringBuilder();
            String line;
            while ((line = request.getReader().readLine()) != null) {
                sb.append(line);
            }
            JSONObject jsonRequest = new JSONObject(sb.toString());

            String username = jsonRequest.getString("username");
            String password = jsonRequest.getString("password");

            // Validate inputs
            if (username == null || password == null || username.isEmpty() || password.isEmpty()) {
                jsonResponse.put("success", false);
                jsonResponse.put("message", "Invalid input");
                out.print(jsonResponse);
                return;
            }

            // Authenticate user with database
            if (DBFunction.authenticateUser(username, password)) {
                jsonResponse.put("success", true);
                jsonResponse.put("message", "Login successful");
            } else {
                jsonResponse.put("success", false);
                jsonResponse.put("message", "Invalid username or password");
            }
        } catch (Exception e) {
            jsonResponse.put("success", false);
            jsonResponse.put("message", "Server error");
            e.printStackTrace();
        }

        out.print(jsonResponse);
        out.flush();
    }

}
