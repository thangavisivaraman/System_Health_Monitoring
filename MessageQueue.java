package system;

import java.util.LinkedList;
import java.util.Queue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;

public class MessageQueue {
	public static BlockingQueue<String> queue = new LinkedBlockingQueue<>();

	  private static boolean clientConnected = false; // Flag to track if client is connected

	    public static void addMessage(String message) {
	        if (clientConnected) { // Only add message if the client is connected
	            queue.add(message);
	        } else {
	            System.out.println("Client not connected. Message discarded.");
	        }
	    }

	    public static String getMessage() {
	        return queue.poll(); // Get and remove the message from the queue
	    }

	    public static void setClientConnected(boolean status) {
	        clientConnected = status;
	    }

	    public static boolean isClientConnected() {
	        return clientConnected;
	    }
}
