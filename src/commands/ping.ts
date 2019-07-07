import { Message } from 'discord.js';
import { Command } from '../models';

function handlePing(message: Message) {
  message.channel.send('Pong! :D');
}

const command: Command = {
  name: 'ping',
  description: 'No surprises here...',
  execute: handlePing,
};

export default command;
