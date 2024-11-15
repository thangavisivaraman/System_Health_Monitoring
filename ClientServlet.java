package system;

import jakarta.servlet.ServletException;
import jakarta.servlet.annotation.WebServlet;
import jakarta.servlet.http.HttpServlet;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.PrintWriter;
import java.util.List;

import com.google.gson.JsonObject;


@WebServlet("/ClientServlet")
public class ClientServlet extends HttpServlet {
	private static final long serialVersionUID = 1L;
   
    public ClientServlet() {
        super();
        // TODO Auto-generated constructor stub
    }

	
	protected void doGet(HttpServletRequest request, HttpServletResponse response) throws ServletException, IOException {	
		
		PrintWriter writer = response.getWriter();
		  response.setContentType("application/json");
		String parameter = request.getParameter("table");
		 int int1 = Integer.parseInt(request.getParameter("offset"));
		 int int2 = Integer.parseInt(request.getParameter("limit"));
		 
		List<JsonObject> dataWithLimitAndOffset = DBFunction.getDataWithLimitAndOffset(parameter, int2, int1);
		 writer.println(dataWithLimitAndOffset.toString());
		 	
		
	}

	
	protected void doPost(HttpServletRequest request, HttpServletResponse response) throws ServletException, IOException {
	
	}

}
