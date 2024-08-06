import http from 'k6/http';
import { sleep } from 'k6';
export const options = {
  vus: 95,
  duration: '1m',
  tags: {
    name: 'url',
  },
};
export default function () {
  http.get('http://localhost:8080/api/v1/authors');
  //sleep(0.01);
  var resp = http.post('http://localhost:8080/api/v1/authors',JSON.stringify({
    name:"qwe1",
    bio: "asdfdsf"
  }),{
    headers: { 'Content-Type': 'application/json' },
  });
  var createdId = resp.json().id
  //sleep(0.01);

  http.get('http://localhost:8080/api/v1/authors/'+createdId);
  //sleep(0.01);
  http.del('http://localhost:8080/api/v1/authors/'+createdId);
  //sleep(0.01);
  http.get('http://localhost:8080/api/v1/authors');
  //sleep(0.01);
}