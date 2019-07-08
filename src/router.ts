import { Collection, Message } from 'discord.js';

interface Command {
    name: string;
    description: string;
    execute: (message: Message, args: string[]) => void;
}

const prefix = "!" // This is temporary
const commands: Collection<string, Command> = new Collection();

// Add a route to the router, specifying the name, description, and what to execute
export function addRoute(name: string, desc: string, exec: (message: Message, args: string[]) => void) {
    commands.set(name, <Command> {
        name: name,
        description: desc,
        execute: exec,
    });
}

// Handle incoming messages
export function route(message: Message) {  
    // Ignore all messages without the prefix or those sent by a bot
    if (!message.content.startsWith(prefix as string) || message.author.bot) {
        return;
    }

    // Build variables based on the command executed
    const args = message.content.slice(prefix!.length).split(/ +/);
    const command = args.shift()!.toLowerCase();

    // If the command does not exist, return
    if (!commands.has(command)) {
        return;
    }

    // Attempt to execute the command
    try {
        const cmd = commands.get(command);
        cmd && cmd.execute(message, args);
    } catch (error) {
        console.error(error);
        message.reply("There was an error while trying to execute that command");
    }
}

// Built-in helper function (literally), used to print commands and descriptions
export function help(message: Message, args: string[]) {
    message.channel.send("This will be a help message eventually.")
}