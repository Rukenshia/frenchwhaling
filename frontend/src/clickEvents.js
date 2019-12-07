import axios from 'axios';

export async function reportClick(type) {
    try {
        await axios.post('https://whaling-api.in.fkn.space/click', type);
    } catch(e) {
        // We don't really care about click event errors to be honest
        console.log(e);
    }
}