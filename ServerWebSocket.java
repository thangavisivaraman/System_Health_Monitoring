package system;

import java.io.IOException;
import java.util.Set;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;

import org.json.JSONObject;

import jakarta.websocket.OnClose;
import jakarta.websocket.OnError;
import jakarta.websocket.OnMessage;
import jakarta.websocket.OnOpen;
import jakarta.websocket.Session;
import jakarta.websocket.server.ServerEndpoint;

@ServerEndpoint("/serverws")
public class ServerWebSocket {

	@OnOpen
	public void onOpen(Session session) throws IOException {
		System.out.println("WebSocket connection opened!");
		session.getBasicRemote().sendText("Hello, WebSocket client!");
	}

	@OnMessage
	public void onMessage(String message, Session session) throws Exception {
		System.out.println("Received message: " + message);

		// Add the received message to the queue
		MessageQueue.addMessage(message);
		JSONObject jsonObject = new JSONObject(message);

		// Get all keys dynamically
		for (String key : jsonObject.keySet()) {
			JSONObject systemInfo = jsonObject.getJSONObject(key);

			DBFunction.insertData(key, systemInfo);
		}
	}

	@OnClose
	public void onClose(Session session) {
		System.out.println("WebSocket connection closed!");
	}

	@OnError
	public void onError(Session session, Throwable throwable) {
		System.err.println("WebSocket error: " + throwable.getMessage());
	}
}
