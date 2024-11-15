package system;

import java.io.IOException;

import jakarta.websocket.OnClose;
import jakarta.websocket.OnError;
import jakarta.websocket.OnMessage;
import jakarta.websocket.OnOpen;
import jakarta.websocket.Session;
import jakarta.websocket.server.ServerEndpoint;

@ServerEndpoint("/clientws")
public class ClientWebSocket {
	
	@OnOpen
	public static void onOpen(Session session) throws IOException {
		System.out.println("WebSocket connection opened in clientws!");
		 // Mark that the client is connected
        MessageQueue.setClientConnected(true);

        while (session.isOpen()) {
            // Simulate processing and getting the message from the queue
            String response = MessageQueue.getMessage();

            if (response == null) {
                response = "No message in the queue.";
            }

            // Send the response back to the client
            session.getBasicRemote().sendText("Message from queue: " + response);
        }

        // Mark that the client is disconnected once the session is closed
        MessageQueue.setClientConnected(false);
	}

	@OnMessage
	public void onMessage(String message, Session session) throws IOException {
		System.out.println("Message received in clientws: " + message);

	}

	@OnClose
	public void onClose(Session session) {
		System.out.println("WebSocket connection closed in clientws!");
	}

	@OnError
	public void onError(Session session, Throwable throwable) {
		System.err.println("WebSocket error: " + throwable.getMessage());
	}
}
